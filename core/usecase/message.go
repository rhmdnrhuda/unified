package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/rhmdnrhuda/unified/core/common"
	"github.com/rhmdnrhuda/unified/core/constant"
	"github.com/rhmdnrhuda/unified/core/entity"
	"github.com/rhmdnrhuda/unified/pkg/logger"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	UniBuddy              = make(map[string][]entity.Message)
	State                 = make(map[string]constant.State)
	UniversityPreferences = make(map[string][]string)
	MajorPreferences      = make(map[string][]string)
	ResetState            = false
	SelectedTalent        = make(map[string]entity.Talent)
	BisonChatRequestMap   = make(map[string]entity.BisonChatRequest)
)

type MessageUseCase struct {
	vertex     VertexOutBound
	ada        AdaOutBound
	cache      Cache
	userRepo   UserRepository
	talentRepo TalentRepository
	alertRepo  AlertRepository
	log        logger.Interface
}

func NewMessageUseCase(vertex VertexOutBound, ada AdaOutBound, cache Cache, userRepo UserRepository,
	talent TalentRepository, alert AlertRepository, log logger.Interface) *MessageUseCase {
	return &MessageUseCase{
		vertex:     vertex,
		ada:        ada,
		cache:      cache,
		userRepo:   userRepo,
		talentRepo: talent,
		alertRepo:  alert,
		log:        log,
	}
}

func (m *MessageUseCase) ProcessMessage(ctx context.Context, req entity.MessageRequest) (string, error) {
	var currentState constant.State
	reqMessage := common.PrepareMessage(req, "", "text")

	if val, ok := State[req.FromNo]; !ok {
		if common.IsIgnore(req.Data.Text) {
			return "", nil
		}

		if common.IsReset(req.Data.Text) {
			m.entryPoint(ctx, req)
			return "", nil
		}

		userTempData, err := m.classifyMessage(ctx, req.Data.Text)
		if err != nil || userTempData == nil {
			m.entryPoint(ctx, req)
			State[req.FromNo] = constant.UNI_CHECK

			return "", err
		}

		if strings.EqualFold(userTempData.Feature, constant.FEATURE_UNI_ALERT) {
			m.processUniAlert(ctx, req, userTempData.UniversityPreferences)
			return "", err
		} else if strings.EqualFold(userTempData.Feature, constant.FEATURE_UNI_BUDDY) {
			UniversityPreferences[req.FromNo] = userTempData.UniversityPreferences
			MajorPreferences[req.FromNo] = userTempData.MajorPreferences
			m.processUniBuddy(context.Background(), req)
			State[req.FromNo] = constant.UNI_BUDDY
			return "", err
		} else if strings.EqualFold(userTempData.Feature, constant.FEATURE_UNI_CONNECT) {
			m.processUniConnect(ctx, req)
			return "", err
		} else {
			m.entryPoint(ctx, req)
			return "", err
		}
	} else {
		currentState = val

		if strings.EqualFold(req.Data.Type, "interactive") {
			m.handleInteractiveMessage(ctx, req)
			return "", nil
		}

		if strings.EqualFold(req.Data.Text, "hi") || strings.EqualFold(req.Data.Text, "hello") {
			m.entryPoint(ctx, req)
			State[req.FromNo] = constant.UNI_CHECK
			return "", nil
		}

		if currentState == constant.UNI_BUDDY {
			m.processUniBuddy(context.Background(), req)
			return "", nil
		} else if currentState == constant.UNI_ALERT {
			UniversityPreferences[req.FromNo] = m.getUniversityFromMessage(ctx, req.Data.Text)
			m.processUniAlert(ctx, req, UniversityPreferences[req.FromNo])
			return "", nil
		} else if currentState == constant.UNI_CHECK {
			m.processUniCheck(ctx, req)
			return "", nil
		} else if currentState == constant.UNI_CONNECT {
			m.processUniConnect(ctx, req)
		}
	}

	return reqMessage.Text, nil
}

func (m *MessageUseCase) processUniCheck(ctx context.Context, req entity.MessageRequest) {
	if _, valid := UniversityPreferences[req.FromNo]; !valid {
		up := m.getUniversityFromMessage(ctx, req.Data.Text)
		if len(up) <= 0 {
			msg := common.PrepareMessage(req, "Ok, no problem! We'll help you figure out what you're looking for 🥂. But before we get started, please answer one more question.", "")
			m.ada.SendMessage(ctx, msg)
		}

		time.Sleep(1 * time.Second)
		UniversityPreferences[req.FromNo] = up

		msg := common.PrepareMessage(req, "What's the major you're interested in pursuing?", "")
		m.ada.SendMessage(ctx, msg)
		return
	}

	mp := m.getMajorFromMessage(ctx, req.Data.Text)
	MajorPreferences[req.FromNo] = mp
	go m.updateUserData(ctx, req, UniversityPreferences[req.FromNo], MajorPreferences[req.FromNo])

	m.nextStateFromUniCheck(ctx, req)
}

func (m *MessageUseCase) processUniAlert(ctx context.Context, req entity.MessageRequest, uniPreferences []string) {
	if len(uniPreferences) <= 0 {
		if State[req.FromNo] == constant.UNI_ALERT {
			reqMessage := common.PrepareMessage(req, "University name is not valid", "")

			err := m.ada.SendMessage(ctx, reqMessage)
			if err != nil {
				m.log.Error("askUniversityPreferences Failed Send Message, err: %v", err)
			}

			return
		}

		universityPreferences, _ := m.getUserPreferences(ctx, req.Data)
		if len(universityPreferences) <= 0 {
			State[req.FromNo] = constant.UNI_ALERT
			m.askUniversityPreferences(ctx, req)
			return
		}

		uniPreferences = universityPreferences
	}

	m.saveAlert(ctx, req, uniPreferences)
}

func (m *MessageUseCase) processUniConnect(ctx context.Context, req entity.MessageRequest) {
	if len(UniversityPreferences[req.FromNo]) <= 0 && len(MajorPreferences[req.FromNo]) <= 0 {
		//reqMessage := common.PrepareMessage(req, fmt.Sprintf("You need to set preferences first"), "")
		//m.ada.SendMessage(ctx, reqMessage)
		m.askUniversityPreferences(ctx, req)
		State[req.FromNo] = constant.UNI_CHECK
		return
	}

	talent, err := m.talentRepo.FindTalentByUniversityAndMajor(ctx, UniversityPreferences[req.FromNo], MajorPreferences[req.FromNo])
	if err != nil {
		reqMessage := common.PrepareMessage(req, fmt.Sprintf("Unfortunately, we couldn't find the perfect mentor based on your preferences."), "")
		m.ada.SendMessage(ctx, reqMessage)
		time.Sleep(1 * time.Second)
		m.endState(ctx, req)
		return
	}

	SelectedTalent[req.FromNo] = talent

	reqMessage := common.PrepareMessage(req, fmt.Sprintf("We can connect you to %s as a %s %s from %s 🏫. Are you interested to have a 45-minutes session to discuss more about your university or major preference with %s? ☎️",
		talent.Name, talent.Major, talent.Status, talent.University, talent.Name), "")
	m.ada.SendMessage(ctx, reqMessage)

	time.Sleep(1 * time.Second)

	paymentURL := fmt.Sprintf("https://unified-payment.vercel.app/payment?phone=%s&price=Rp.50.000", req.FromNo)
	reqMessage = common.PrepareMessage(req, "Please continue to payment using the link below if you are interested to have a session 👥\n\n "+paymentURL, "")
	m.ada.SendMessage(ctx, reqMessage)
	deleteAllCache(req.FromNo)
}

func (m *MessageUseCase) saveAlert(ctx context.Context, req entity.MessageRequest, uniPreferences []string) error {
	var events []entity.Event
	var err error
	listAlert := "Here are some important dates you can take notes📝:\n"

	var dataAlert []entity.Alert

	for _, uni := range uniPreferences {
		var response *entity.BisonTextResponse
		request := provideBisonTextRequest(fmt.Sprintf(constant.TemplatePromptGetUniTimeline, uni), 2048)
		response, err = m.vertex.DoCallVertexAPIText(ctx, request, m.getAccessToken())
		if err != nil {
			m.log.Error("failed call DoCallVertexAPIText processUniAlert, request: %v\nerror: %v", request, err)
			continue
		}

		var alert entity.AlertResponse
		content := strings.ReplaceAll(response.Predictions[0].Content, "JSON", "")
		content = strings.ReplaceAll(content, "`", "")
		err = json.Unmarshal([]byte(content), &alert)
		if err != nil {
			var ev []entity.Event
			err = json.Unmarshal([]byte(content), &ev)
			if err != nil {
				m.log.Error("failed call Unmarshal, data: %v\nerror: %v", content, err)
				continue
			}
			alert.Events = ev
		}

		for i, ev := range alert.Events {
			date := common.ProcessDate(ev.Date)
			listAlert = fmt.Sprintf("%s\n%d. %s - %s", listAlert, i+1, ev.EventTitle, date.Format("02 January 2006"))
			dataAlert = append(dataAlert, entity.Alert{
				UserID:     req.FromNo,
				Date:       date.Unix(),
				Message:    ev.EventTitle,
				University: uni,
			})
		}

		events = append(events, alert.Events...)
	}

	if err != nil {
		message := fmt.Sprintf("I am so sorry for the inconvenience, but I am currently experiencing some technical difficulties. \nMy team is working hard to resolve the issue as quickly as possible. In the meantime, please try again later. Thank you for your patience and understanding.")
		reqMessage := common.PrepareMessage(req, message, "")
		err = m.ada.SendMessage(ctx, reqMessage)
		m.endState(ctx, req)
		return err
	}

	message := fmt.Sprintf("Sure thing! We've already set up reminders to keep you in the loop about important events related to your registration timeline at %s 📅 ", strings.Join(uniPreferences, "\n"))
	reqMessage := common.PrepareMessage(req, message, "")
	err = m.ada.SendMessage(ctx, reqMessage)
	if err != nil {
		m.log.Error("askUniversityPreferences Failed Send Message, err: %v", err)
	}

	if len(events) > 0 {
		time.Sleep(500 * time.Millisecond)
		reqMessage = common.PrepareMessage(req, listAlert, "")
		m.ada.SendMessage(ctx, reqMessage)

		go m.alertRepo.Create(context.Background(), dataAlert)
	}

	time.Sleep(500 * time.Millisecond)
	sticker := common.PrepareStickerMessage(req, "119bba90-da64-469c-92a2-7ccc76046618")
	m.ada.SendMessage(ctx, sticker)

	time.Sleep(500 * time.Millisecond)
	m.endState(ctx, req)

	return nil
}

func (m *MessageUseCase) askUniversityPreferences(ctx context.Context, req entity.MessageRequest) error {
	reqMessage := common.PrepareMessage(req, "Which university are you dreaming of going to? 😊", "")
	err := m.ada.SendMessage(ctx, reqMessage)
	if err != nil {
		m.log.Error("askUniversityPreferences Failed Send Message, err: %v", err)
	}

	return err
}

func (m *MessageUseCase) getUniversityFromMessage(ctx context.Context, message string) []string {
	request := provideBisonTextRequest(fmt.Sprintf(constant.TemplateValidateUniversityName, message), 200)
	response, err := m.vertex.DoCallVertexAPIText(ctx, request, m.getAccessToken())
	if err != nil {
		m.log.Error("failed call DoCallVertexAPIText, request: %v\nerror: %v", request, err)
		return []string{}
	}

	content := strings.ReplaceAll(response.Predictions[0].Content, " ", "")
	if strings.EqualFold(content, "no") {
		return []string{}
	}

	ctn := strings.ReplaceAll(response.Predictions[0].Content, "\n", "")
	ctn = strings.ReplaceAll(ctn, "`", "")
	ctn = strings.ReplaceAll(ctn, "pascal", "")
	return strings.Split(strings.TrimSpace(ctn), ",")
}

func (m *MessageUseCase) getMajorFromMessage(ctx context.Context, message string) []string {
	request := provideBisonTextRequest(fmt.Sprintf(constant.TemplateValidateMajor, message), 200)
	response, err := m.vertex.DoCallVertexAPIText(ctx, request, m.getAccessToken())
	if err != nil {
		m.log.Error("failed call DoCallVertexAPIText, request: %v\nerror: %v", request, err)
		return []string{}
	}

	content := strings.ReplaceAll(response.Predictions[0].Content, " ", "")
	if strings.EqualFold(content, "no") {
		return []string{}
	}

	return strings.Split(strings.TrimSpace(response.Predictions[0].Content), ", ")
}

func (m *MessageUseCase) classifyMessage(ctx context.Context, message string) (*entity.UserTemporaryData, error) {
	request := provideBisonTextRequest(fmt.Sprintf(constant.TemplatePromptClassify, message), 200)
	response, err := m.vertex.DoCallVertexAPIText(ctx, request, m.getAccessToken())
	if err != nil {
		m.log.Error("failed call DoCallVertexAPIText, request: %v\nerror: %v", request, err)
		return nil, err
	}

	var userTempData *entity.UserTemporaryData
	content := strings.ReplaceAll(response.Predictions[0].Content, "JSON", "")
	content = strings.ReplaceAll(content, "`", "")
	err = json.Unmarshal([]byte(content), &userTempData)
	if err != nil {
		m.log.Error("failed call Unmarshal, data: %v\nerror: %v", content, err)
		return nil, err
	}

	return userTempData, nil
}

func (m *MessageUseCase) processUniBuddy(ctx context.Context, req entity.MessageRequest) {
	bisonChatReq := entity.BisonChatRequest{}
	if val, ok := BisonChatRequestMap[req.FromNo]; !ok {
		bisonChatReq = initBisonChatUniBuddyRequest(UniversityPreferences[req.FromNo], MajorPreferences[req.FromNo])
	} else {
		bisonChatReq = val
	}

	var messages []entity.Message
	if val, ok := UniBuddy[req.FromNo]; ok {
		messages = val
		messages = append(messages, entity.Message{
			Author:  "user",
			Content: req.Data.Text,
		})
	} else {
		content := "Give descriptive information of top 5 universities in the world and top 5 majors in the world. Also ask user whether they have any question regarding the recommendation or user wants to be guided with some questions to help on picking the best university or major for them. in one message"
		_, ok := State[req.FromNo]
		if !ok {
			content = req.Data.Text
		} else {
			if len(MajorPreferences[req.FromNo]) > 0 && len(UniversityPreferences[req.FromNo]) > 0 {
				content = fmt.Sprintf("Give information about %s and %s in long informal description. Also ask user whether they have any question regarding the recommendation or wants to know about other universities or major.", strings.Join(MajorPreferences[req.FromNo], ", "), strings.Join(UniversityPreferences[req.FromNo], ", "))
			} else if len(MajorPreferences[req.FromNo]) > 0 {
				content = fmt.Sprintf("Give descriptive information of top 5 universities in the world for %s major and describe %s in long informal description. Also ask user whether they have any question regarding the recommendation or user wants to be guided with some questions to help on picking the best university for them in one message", strings.Join(MajorPreferences[req.FromNo], ", "), strings.Join(MajorPreferences[req.FromNo], ", "))
			} else if len(UniversityPreferences[req.FromNo]) > 0 {
				content = fmt.Sprintf("Give descriptive information of top 5 major in the %s for and describe %s in long informal description. Also ask user whether they have any question regarding the recommendation or user wants to be guided with some questions to help on picking the best university for them in one message", strings.Join(UniversityPreferences[req.FromNo], ", "), strings.Join(UniversityPreferences[req.FromNo], ", "))
			}
		}

		messages = append(messages, entity.Message{
			Author:  "user",
			Content: content,
		})
	}

	fmt.Printf("user: %s\ncurrent message uni-buddy: %+v\n", req.FromNo, messages)

	bisonChatReq.Instances[0].Messages = messages
	res, err := m.vertex.DoCallVertexAPIChat(ctx, bisonChatReq, m.getAccessToken())
	if err != nil {
		m.log.Error("DoCallVertexAPIChat error: %v\nmessage: %v", err, messages)
		return
	}

	newMessage := res.Predictions[0].Candidates[0]
	if !strings.EqualFold("bot", newMessage.Author) {
		newMessage.Author = "bot"
	}
	messages = append(messages, newMessage)
	UniBuddy[req.FromNo] = messages

	if len(messages) > 10 {
		//
	}
	response := entity.UniBuddyResponse{}
	message := res.Predictions[0].Candidates[0].Content
	content := strings.ReplaceAll(res.Predictions[0].Candidates[0].Content, "\t", " ")
	content = strings.ReplaceAll(message, "\n", "\\n")
	err = json.Unmarshal([]byte(content), &response)
	if err != nil {
		msg := common.PrepareMessage(req, message, "")
		m.ada.SendMessage(ctx, msg)
		m.log.Error("error unmarshal: %v\nerr: %v", message, err)
		return
	}

	message = response.Message
	go m.updateUserData(ctx, req, response.University, response.Major)
	msg := common.PrepareMessage(req, message, "")
	m.ada.SendMessage(ctx, msg)

	if response.IsFinished {
		time.Sleep(500 * time.Millisecond)
		m.nextStepUniBuddy(context.Background(), req)
		return
	}

	return
}

func (m *MessageUseCase) updateUserData(ctx context.Context, req entity.MessageRequest, uniPreferences, majorPreferences []string) {
	needUpdateDB := false

	if len(uniPreferences) > 0 {
		needUpdateDB = true
		UniversityPreferences[req.FromNo] = uniPreferences
	}

	if len(majorPreferences) > 0 {
		needUpdateDB = true
		MajorPreferences[req.FromNo] = majorPreferences
	}

	go func(needUpdate bool) {
		if needUpdate {
			userData := &entity.User{
				Name:                  req.Data.CustName,
				Number:                req.FromNo,
				UniversityPreferences: UniversityPreferences[req.FromNo],
				MajorPreferences:      MajorPreferences[req.FromNo],
			}

			err := m.userRepo.Update(context.Background(), userData)
			if err != nil {
				m.log.Error("Failed to update DB, data: %v, error: %v", userData, err)
			}
		}
	}(needUpdateDB)
}

func (m *MessageUseCase) nextStepUniBuddy(ctx context.Context, req entity.MessageRequest) {
	talent, err := m.talentRepo.FindTalentByUniversityAndMajor(ctx, UniversityPreferences[req.FromNo], MajorPreferences[req.FromNo])
	if err == nil && talent.Name != "" {
		message := fmt.Sprintf("Would you like to reach out to folks from your dream university or major? 🎓")
		msg := common.PrepareMessageButton(req, message, "", "", constant.ButtonYesOrNo)
		m.ada.SendMessageButton(ctx, msg)

		time.Sleep(500 * time.Millisecond)
		sticker := common.PrepareStickerMessage(req, "5a35c3fb-d54c-4e53-b4be-81961ab9137a")
		m.ada.SendMessage(ctx, sticker)
		State[req.FromNo] = constant.UNI_BUDDY_TO_UNI_CONNECT
		return
	}

	message := fmt.Sprintf("Do you want UNIFIED to send you alert regarding %s important event dates? ☎️", strings.Join(UniversityPreferences[req.FromNo], ", "))
	msg := common.PrepareMessageButton(req, message, "", "", constant.ButtonYesOrNo)
	m.ada.SendMessageButton(ctx, msg)

	State[req.FromNo] = constant.UNI_BUDDY_TO_UNI_ALERT
}

func (m *MessageUseCase) nextStateFromUniCheck(ctx context.Context, req entity.MessageRequest) {
	message := ""
	var btn []string

	if len(MajorPreferences[req.FromNo]) <= 0 && len(UniversityPreferences[req.FromNo]) <= 0 {
		go m.processUniBuddy(context.Background(), req)
		State[req.FromNo] = constant.UNI_BUDDY
		return
	} else if len(UniversityPreferences[req.FromNo]) > 0 {
		message = fmt.Sprintf("It's fantastic that you're interested in pursuing your studies at %s! How can UNIFIED assist you with your plan? 🚀\n1. 🎓Uni-Buddy:  Ask us a question, and we will provide personalized information about studying for you\n2. 👥 Uni-Connect: Get in touch with someone from your dream university or major!\n3. ☎️ Uni-Alert: Let's set up alerts for you regarding important university timelines.", strings.Join(UniversityPreferences[req.FromNo], ", "))
		if len(MajorPreferences[req.FromNo]) > 0 {
			message = fmt.Sprintf("It's fantastic that you're interested in pursuing a %s at %s! How can UNIFIED assist you with your plans? 🚀\n1. 🎓Uni-Buddy:  Ask us a question, and we will provide personalized information about studying for you\n2. 👥 Uni-Connect: Get in touch with someone from your dream university or major!\n3. ☎️ Uni-Alert: Let's set up alerts for you regarding important university timelines.", strings.Join(MajorPreferences[req.FromNo], ", "), strings.Join(UniversityPreferences[req.FromNo], ", "))
		}
		btn = constant.ButtonAllFeature
	} else if len(MajorPreferences[req.FromNo]) > 0 {
		message = fmt.Sprintf("It's great to hear that you're interested in pursuing %s! How can UNIFIED assist you with your plan? 🚀\n1. 🎓Uni-Buddy:  Ask us a question, and we will provide personalized information about studying for you\n2. 👥 Uni-Connect: Get in touch with someone from your dream university or major!", strings.Join(MajorPreferences[req.FromNo], ", "))
		btn = constant.ButtonUniBuddyUniConnect
	}

	msg := common.PrepareMessageButton(req, message, "", "", btn)
	m.ada.SendMessageButton(ctx, msg)
	time.Sleep(500 * time.Millisecond)
	sticker := common.PrepareStickerMessage(req, "8e2d7746-2585-4a37-a0c5-e0d54841ceb4")
	m.ada.SendMessage(ctx, sticker)
}

func (m *MessageUseCase) getUserPreferences(ctx context.Context, user entity.DataRequest) ([]string, []string) {
	userData, err := m.userRepo.FindUserByNumber(ctx, user.CustNo)
	if err != nil {
		go m.userRepo.Create(context.Background(), &entity.User{
			Name:                  user.CustName,
			Number:                user.CustNo,
			UniversityPreferences: nil,
			MajorPreferences:      nil,
		})
		return nil, nil
	}

	return userData.UniversityPreferences, userData.MajorPreferences
}

func initBisonChatUniBuddyRequest(uniPreferences, majorPreferences []string) entity.BisonChatRequest {
	context := constant.ContextUniBuddyUniNoMajorNo
	if len(uniPreferences) > 0 && len(majorPreferences) > 0 {
		context = fmt.Sprintf(constant.ContextUniBuddyUniYesMajorYes)
	} else if len(uniPreferences) > 0 {
		context = fmt.Sprintf(constant.ContextUniBuddyUniYesMajorNo)
	} else if len(majorPreferences) > 0 {
		context = fmt.Sprintf(constant.ContextUniBuddyUniNoMajorYes)
	}

	return entity.BisonChatRequest{
		Instances: []entity.Instance{
			{
				Context:  context,
				Examples: constant.ExampleBisonChatUniBuddy,
			},
		},
		Parameters: entity.Parameter{
			Temperature:     0.3,
			MaxOutputTokens: 2048,
			TopP:            0.8,
			TopK:            40,
		},
	}
}

func (m *MessageUseCase) handleInteractiveMessage(ctx context.Context, req entity.MessageRequest) {
	if req.Data.Text == constant.ButtonUniAlert {
		m.processUniAlert(ctx, req, UniversityPreferences[req.FromNo])
		return
	} else if req.Data.Text == constant.ButtonUniBuddy {
		m.processUniBuddy(ctx, req)
		State[req.FromNo] = constant.UNI_BUDDY
		return
	} else if req.Data.Text == constant.ButtonUniConnect {
		m.processUniConnect(ctx, req)
		return
	}

	if State[req.FromNo] == constant.UNI_BUDDY_TO_UNI_ALERT {
		if strings.EqualFold(req.Data.Text, "yes") {
			m.processUniAlert(ctx, req, UniversityPreferences[req.FromNo])
		}

		m.endState(ctx, req)
	} else if State[req.FromNo] == constant.UNI_BUDDY_TO_UNI_CONNECT {
		if strings.EqualFold(req.Data.Text, "yes") {
			m.processUniConnect(ctx, req)
		} else {
			m.endState(ctx, req)
		}

	} else if ResetState {
		if strings.EqualFold(req.Data.Text, "yes") {
			reqMessage := common.PrepareMessage(req, "Okay..  How can I assist you today? 😊", "")
			m.ada.SendMessage(ctx, reqMessage)
			deleteAllCache(req.FromNo)
		} else {

		}
	}
}

func provideBisonTextRequest(prompt string, maxOutput float64) entity.BisonTextRequest {
	return entity.BisonTextRequest{
		Instances: []entity.InstanceBisonText{
			{
				Prompt: prompt,
			},
		},
		Parameters: entity.Parameter{
			Temperature:     0.2,
			MaxOutputTokens: maxOutput,
			TopP:            0.8,
			TopK:            40,
		},
	}
}

func (m *MessageUseCase) getAccessToken() string {
	var token string
	token = m.cache.Get(constant.RedisAccessTokenKey)
	if token != "" {
		return token
	}

	cmd := exec.Command("gcloud", "auth", "print-access-token")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if len(string(output)) < 2 {
		return ""
	}

	token = string(output)[:len(string(output))-2]
	go m.cache.Set(constant.RedisAccessTokenKey, token, constant.RedisAccessTokenTTL)

	return token
}

func (m *MessageUseCase) entryPoint(ctx context.Context, req entity.MessageRequest) {
	reqMessage := common.PrepareMessage(req, fmt.Sprintf("Hello %s! 👋 I'm Unified, your trusty student personal assistant. My mission is to make your journey to choosing a bachelor's university and major as smooth as possible.", req.Data.CustName), "")
	m.ada.SendMessage(ctx, reqMessage)
	time.Sleep(1 * time.Second)
	reqMessage.Text = "Feel free to ask me anything, from university recommendations to major insights—I'm here to help you every step of the way! 😊🎓"
	m.ada.SendMessage(ctx, reqMessage)

	time.Sleep(500 * time.Millisecond)
	sticker := common.PrepareStickerMessage(req, "2cc41558-02c7-406c-8704-b569ee775a88")
	m.ada.SendMessage(ctx, sticker)

	time.Sleep(1 * time.Second)
	m.askUniversityPreferences(ctx, req)
	deleteAllCache(req.FromNo)
	State[req.FromNo] = constant.UNI_CHECK
}

func deleteAllCache(no string) {
	delete(UniversityPreferences, no)
	delete(MajorPreferences, no)
	delete(UniBuddy, no)
	delete(State, no)
}

func (m *MessageUseCase) endState(ctx context.Context, req entity.MessageRequest) {
	reqMessage := common.PrepareMessage(req, "If you need any assistance, please don't hesitate to contact us. Simply greet me to start chatting with me again. Have a wonderful day!", "")
	m.ada.SendMessage(ctx, reqMessage)
	deleteAllCache(req.FromNo)
}

func (m *MessageUseCase) PaymentCallback(ctx context.Context, phone string) {
	req := entity.MessageRequest{
		FromNo:      phone,
		Platform:    "WA",
		AccountNo:   "60136958751",
		AccountName: "UNIFIED",
		Data:        entity.DataRequest{},
	}

	msg := common.PrepareMessage(req, "Payment completed\nplease click link below to set your schedule\n"+SelectedTalent[phone].CalendarURL, "template")

	msg.TemplateName = "set_schedulue"
	m.ada.SendMessage(ctx, msg)

	time.Sleep(500 * time.Millisecond)

	sticker := common.PrepareStickerMessage(req, "67477312-edec-4ec1-8940-38678f2dea57")
	m.ada.SendMessage(ctx, sticker)
	time.Sleep(1 * time.Second)

	msg = common.PrepareMessage(req, "Have a fantastic day! 🎉 If you ever need assistance in the future, don't hesitate to reach out. Take care!", "")
	m.ada.SendMessage(ctx, msg)
}

func (m *MessageUseCase) RunCron(ctx context.Context) {
	req := entity.MessageRequest{
		FromNo:      "6282122277701",
		Platform:    "WA",
		AccountName: "UNIFIED",
		AccountNo:   "60136958751",
		Data:        entity.DataRequest{},
	}

	res, err := m.alertRepo.FindAlert(ctx, 365)
	if err != nil {
		return
	}

	now := time.Now()
	for _, v := range res {
		req.FromNo = v.UserID
		unixTimestamp := v.Date
		message := "🚨 Don't forget about today's Document Application Deadline for Stanford University!🚨"
		if !time.Unix(unixTimestamp, 0).Equal(now) {
			daysDiff := int(time.Unix(unixTimestamp, 0).Sub(now).Hours() / 24)
			message = fmt.Sprintf("%d more days towards %s for %s! Don't forget to prepare 😉", daysDiff, v.Messages[0], v.University)
		}

		msg := common.PrepareMessage(req, message, "")
		m.ada.SendMessage(ctx, msg)
	}
}

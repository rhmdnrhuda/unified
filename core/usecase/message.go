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
		if strings.EqualFold(req.Data.Text, "hi") || strings.EqualFold(req.Data.Text, "hello") {
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
			State[req.FromNo] = constant.UNI_ALERT
			return "", err
		} else if strings.EqualFold(userTempData.Feature, constant.FEATURE_UNI_BUDDY) {
			UniversityPreferences[req.FromNo] = userTempData.UniversityPreferences
			MajorPreferences[req.FromNo] = userTempData.MajorPreferences
			m.processUniBuddy(context.Background(), req)
			State[req.FromNo] = constant.UNI_BUDDY
			return "", err
		} else if strings.EqualFold(userTempData.Feature, constant.FEATURE_UNI_CONNECT) {
			m.processUniConnect(ctx, req)
			State[req.FromNo] = constant.UNI_CONNECT
			return "", err
		} else {
			reqMessage.Text = fmt.Sprintf("Hello %s! 👋 I'm Unified, your trusty student personal assistant. My mission is to make your journey to choosing a bachelor's university and major as smooth as possible.", req.Data.CustName)
			err := m.ada.SendMessage(ctx, reqMessage)
			reqMessage.Text = "Feel free to ask me anything, from university recommendations to major insights—I'm here to help you every step of the way! 😊🎓"
			err = m.ada.SendMessage(ctx, reqMessage)
			m.askUniversityPreferences(ctx, req)
			State[req.FromNo] = constant.UNI_CHECK
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
			msg := common.PrepareMessage(req, "[hardcoded]Ok, no problem! We'll help you figure out what you're looking for 🥂. But before we get started, please help me answer one more question", "")
			m.ada.SendMessage(ctx, msg)
		}

		UniversityPreferences[req.FromNo] = up

		msg := common.PrepareMessage(req, "[hardcoded]Do you have a *major* in mind that you're interested in?", "")
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
			reqMessage := common.PrepareMessage(req, "[hardcoded]Your input is not valid", "")

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

	message := fmt.Sprintf("Sure thing! We've already set up reminders to keep you in the loop about important events related to your registration timeline at %s", strings.Join(uniPreferences, "\n"))

	reqMessage := common.PrepareMessage(req, message, "")

	err := m.ada.SendMessage(ctx, reqMessage)
	if err != nil {
		m.log.Error("askUniversityPreferences Failed Send Message, err: %v", err)
	}
}

func (m *MessageUseCase) processUniConnect(ctx context.Context, req entity.MessageRequest) {
	if len(UniversityPreferences[req.FromNo]) <= 0 && len(MajorPreferences[req.FromNo]) <= 0 {
		reqMessage := common.PrepareMessage(req, fmt.Sprintf("[hardcoded] You need to set preferences first"), "")
		m.ada.SendMessage(ctx, reqMessage)
		m.askUniversityPreferences(ctx, req)
		State[req.FromNo] = constant.UNI_CHECK
	}

	talent, err := m.talentRepo.FindTalentByUniversityAndMajor(ctx, UniversityPreferences[req.FromNo], MajorPreferences[req.FromNo])
	if err != nil {
		reqMessage := common.PrepareMessage(req, fmt.Sprintf("[hardcoded]Unortunately, we can't found perfect talent based on your preferences"), "")
		m.ada.SendMessage(ctx, reqMessage)
		delete(State, req.FromNo)
		return
	}

	SelectedTalent[req.FromNo] = talent

	reqMessage := common.PrepareMessage(req, fmt.Sprintf("[hardcoded] You can discuss more about your preferences with: %s from %s with major %s as %s", talent.Name, talent.University, talent.Major, talent.Status), "")
	m.ada.SendMessage(ctx, reqMessage)

	paymentURL := fmt.Sprintf("https://unified-payment.vercel.app/payment?phone=%s&price=Rp.50.000", req.FromNo)
	reqMessage = common.PrepareMessage(req, "[hardcoded] click link below to make a payment if you want to connect.\n "+paymentURL, "")
	m.ada.SendMessage(ctx, reqMessage)
	delete(State, req.FromNo)
}

func (m *MessageUseCase) askUniversityPreferences(ctx context.Context, req entity.MessageRequest) error {
	reqMessage := common.PrepareMessage(req, "[hardcoded] Do you have a university in mind that you're interested in?", "")
	err := m.ada.SendMessage(ctx, reqMessage)
	if err != nil {
		m.log.Error("askUniversityPreferences Failed Send Message, err: %v", err)
	}

	return err
}

func (m *MessageUseCase) getUniversityFromMessage(ctx context.Context, message string) []string {
	request := provideBisonTextRequest(fmt.Sprintf(constant.TemplateValidateUniversityName, message))
	response, err := m.vertex.DoCallVertexAPIText(ctx, request, m.getAccessToken())
	if err != nil {
		m.log.Error("failed call DoCallVertexAPIText, request: %v\nerror: %v", request, err)
		return []string{}
	}

	content := strings.ReplaceAll(response.Predictions[0].Content, " ", "")
	if strings.EqualFold(content, "no") {
		return []string{}
	}

	return strings.Split(strings.TrimSpace(response.Predictions[0].Content), ",")
}

func (m *MessageUseCase) getMajorFromMessage(ctx context.Context, message string) []string {
	request := provideBisonTextRequest(fmt.Sprintf(constant.TemplateValidateMajor, message))
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
	request := provideBisonTextRequest(fmt.Sprintf(constant.TemplatePromptBisonChat, message))
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
	bisonChatReq := initBisonChatUniBuddyRequest(UniversityPreferences[req.FromNo], MajorPreferences[req.FromNo])
	var messages []entity.Message
	if val, ok := UniBuddy[req.FromNo]; ok {
		messages = val
		messages = append(messages, entity.Message{
			Author:  "user",
			Content: req.Data.Text,
		})
	} else {
		content := "start"
		_, ok := State[req.FromNo]
		if !ok {
			content = req.Data.Text
		} else {
			if len(MajorPreferences[req.FromNo]) > 0 && len(UniversityPreferences[req.FromNo]) > 0 {
				content = fmt.Sprintf("Hi Unified, I'm interested to study %s in %s Can you tell me more about it?", strings.Join(MajorPreferences[req.FromNo], ", "), strings.Join(MajorPreferences[req.FromNo], ", "))
			} else if len(MajorPreferences[req.FromNo]) > 0 {
				content = fmt.Sprintf("Hi Unified, I'm interested to study %s major. Can you tell me more about it? Or give me University recommendation from that major", strings.Join(MajorPreferences[req.FromNo], ", "))
			} else {
				content = fmt.Sprintf("Hi Unified, I'm interested to study in %s. Can you tell me more about it? Or give me Major recommendation from that university", strings.Join(UniversityPreferences[req.FromNo], ", "))
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

	response := entity.UniBuddyResponse{}
	message := res.Predictions[0].Candidates[0].Content
	err = json.Unmarshal([]byte(res.Predictions[0].Candidates[0].Content), &response)
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
		go m.nextStepUniBuddy(context.Background(), req)
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
		State[req.FromNo] = constant.UNI_BUDDY_TO_UNI_CONNECT
		return
	}

	message := fmt.Sprintf("[hardcoded]Do you want to get notification from %s", strings.Join(UniversityPreferences[req.FromNo], ", "))
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
		message = fmt.Sprintf("Great to know that you are interested in studying at %s! How can UNIFIED assist you with your plans? 🚀", strings.Join(UniversityPreferences[req.FromNo], ", "))
		if len(MajorPreferences[req.FromNo]) > 0 {
			message = fmt.Sprintf("It's fantastic that you're interested in pursuing a %s at %s! How can UNIFIED assist you with your plans? 🚀", strings.Join(MajorPreferences[req.FromNo], ", "), strings.Join(UniversityPreferences[req.FromNo], ", "))
		}
		btn = constant.ButtonAllFeature
	} else if len(MajorPreferences[req.FromNo]) > 0 {
		message = fmt.Sprintf("Great to know that you are interested in studying %s! How can UNIFIED assist you with your plans? 🚀", strings.Join(MajorPreferences[req.FromNo], ", "))
		btn = constant.ButtonUniBuddyUniConnect
	}

	msg := common.PrepareMessageButton(req, message, "", "", btn)
	m.ada.SendMessageButton(ctx, msg)
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
		context = fmt.Sprintf(constant.ContextUniBuddyUniYesMajorYes, strings.Join(uniPreferences, ", "), strings.Join(majorPreferences, ", "), strings.Join(uniPreferences, ", "), strings.Join(majorPreferences, ", "))
	} else if len(uniPreferences) > 0 {
		context = fmt.Sprintf(constant.ContextUniBuddyUniYesMajorNo, strings.Join(uniPreferences, ", "), strings.Join(uniPreferences, ", "), strings.Join(uniPreferences, ", "), strings.Join(uniPreferences, ", "))
	} else if len(majorPreferences) > 0 {
		context = fmt.Sprintf(constant.ContextUniBuddyUniNoMajorYes, strings.Join(majorPreferences, ", "), strings.Join(majorPreferences, ", "), strings.Join(majorPreferences, ", "), strings.Join(majorPreferences, ", "))
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
		State[req.FromNo] = constant.UNI_ALERT
		return
	} else if req.Data.Text == constant.ButtonUniBuddy {
		m.processUniBuddy(ctx, req)
		State[req.FromNo] = constant.UNI_BUDDY
		return
	} else if req.Data.Text == constant.ButtonUniConnect {
		m.processUniConnect(ctx, req)
		State[req.FromNo] = constant.UNI_CONNECT
		return
	}

	if State[req.FromNo] == constant.UNI_BUDDY_TO_UNI_ALERT {
		if strings.EqualFold(req.Data.Text, "yes") {
			m.processUniAlert(ctx, req, UniversityPreferences[req.FromNo])

			reqMessage := common.PrepareMessage(req, "[hardcoded] Anything else?", "")
			m.ada.SendMessage(ctx, reqMessage)
		} else {
			//	todo: tawarin uni connect
		}
	} else if State[req.FromNo] == constant.UNI_BUDDY_TO_UNI_CONNECT {
		if strings.EqualFold(req.Data.Text, "yes") {
			m.processUniConnect(ctx, req)
			State[req.FromNo] = constant.UNI_CONNECT
		} else {
			//	todo: tawarin uni connect
		}
	} else if ResetState {
		if strings.EqualFold(req.Data.Text, "yes") {
			reqMessage := common.PrepareMessage(req, "Okay..  How can I assist you today? 😊", "")
			m.ada.SendMessage(ctx, reqMessage)
			delete(State, req.FromNo)
		} else {

		}
	}
}

func provideBisonTextRequest(prompt string) entity.BisonTextRequest {
	return entity.BisonTextRequest{
		Instances: []entity.InstanceBisonText{
			{
				Prompt: prompt,
			},
		},
		Parameters: entity.Parameter{
			Temperature:     0.3,
			MaxOutputTokens: 200,
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
	time.Sleep(1 * time.Second)
	m.askUniversityPreferences(ctx, req)
	delete(UniversityPreferences, req.FromNo)
	delete(MajorPreferences, req.FromNo)
	State[req.FromNo] = constant.UNI_CHECK
}

func (m *MessageUseCase) PaymentCallback(ctx context.Context, phone string) {
	msg := common.PrepareMessage(entity.MessageRequest{
		FromNo:      phone,
		Platform:    "WA",
		AccountNo:   "60136958751",
		AccountName: "UNIFIED",
		Data:        entity.DataRequest{},
	}, "Payment completed\nplease click link below to set your schedule\n"+SelectedTalent[phone].CalendarURL, "template")

	msg.TemplateName = "schedule"
	m.ada.SendMessage(ctx, msg)
}

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
)

var (
	UniBuddy              = make(map[string][]entity.Message)
	State                 = make(map[string]constant.State)
	UniversityPreferences = make(map[string][]string)
	MajorPreferences      = make(map[string][]string)
)

type MessageUseCase struct {
	vertex   VertexOutBound
	ada      AdaOutBound
	cache    Cache
	userRepo UserRepository
	log      logger.Interface
}

func NewMessageUseCase(vertex VertexOutBound, ada AdaOutBound, cache Cache, userRepo UserRepository, log logger.Interface) *MessageUseCase {
	return &MessageUseCase{
		vertex:   vertex,
		ada:      ada,
		cache:    cache,
		userRepo: userRepo,
		log:      log,
	}
}

func (m *MessageUseCase) ProcessMessage(ctx context.Context, req entity.MessageRequest) (string, error) {
	var currentState constant.State
	reqMessage := common.PrepareMessage(req, "", "text")

	if val, ok := State[req.FromNo]; !ok {
		if strings.EqualFold(req.Data.Text, "hi") || strings.EqualFold(req.Data.Text, "hello") {
			reqMessage.Text = fmt.Sprintf("Hello %s! 👋 I'm Unified, your trusty student personal assistant. My mission is to make your journey to choosing a bachelor's university and major as smooth as possible.", req.Data.CustName)
			err := m.ada.SendMessage(ctx, reqMessage)
			reqMessage.Text = "Feel free to ask me anything, from university recommendations to major insights—I'm here to help you every step of the way! 😊🎓"
			err = m.ada.SendMessage(ctx, reqMessage)
			m.askUniversityPreferences(ctx, req)
			State[req.FromNo] = constant.UNI_CHECK
			return "", err
		}

		userTempData, err := m.classifyMessage(ctx, req.Data.Text)
		if err != nil || userTempData == nil {
			reqMessage.Text = fmt.Sprintf("Hi %s, Welcome to UNIFIED.", req.Data.CustName)
			err = m.ada.SendMessage(ctx, reqMessage)
			if err != nil {
				m.log.Error("Failed Send Message, err: %v\n userTempData: %v", err, userTempData)
			}

			return "", err
		}

		if strings.EqualFold(userTempData.Feature, constant.FEATURE_UNI_ALERT) {
			m.processUniAlert(ctx, req, userTempData.UniversityPreferences)
			return "", err
		} else if strings.EqualFold(userTempData.Feature, constant.FEATURE_UNI_BUDDY) {
			if len(userTempData.UniversityPreferences) <= 0 || len(userTempData.MajorPreferences) <= 0 {
				userTempData.UniversityPreferences, userTempData.MajorPreferences = m.getUserPreferences(ctx, req.Data)
			}
			m.log.Info("msauk ke uni buddy")
			return "", err
		} else if strings.EqualFold(userTempData.Feature, constant.FEATURE_UNI_CONNECT) {
			m.log.Info("msauk ke uni connect")
			return "", err
		} else {
			m.log.Info("msauk ke uni check")
			return "", err
		}
	} else {
		currentState = val
		if strings.EqualFold(req.Data.Text, "hi") || strings.EqualFold(req.Data.Text, "hello") {
			reqButton := common.PrepareMessageButton(req, "Do you want to reset your previous state?", "", "", []string{"Yes", "No"})
			return "", m.ada.SendButton(ctx, reqButton)
		}
		if currentState == constant.UNI_BUDDY {
			//newReqMessage = m.processUniBuddy(ctx, req)
		} else if currentState == constant.UNI_ALERT {
			UniversityPreferences[req.FromNo] = m.getUniversityFromMessage(ctx, req.Data.Text)
			m.processUniAlert(ctx, req, UniversityPreferences[req.FromNo])
			return "", nil
		} else if currentState == constant.UNI_CHECK {
			m.processUniCheck(ctx, req)
			return "", nil
		} else if currentState == constant.UNI_CONNECT {
			m.prosesUniConnect(ctx, req)
		}
	}

	//go func() {
	//	err := m.ada.SendMessage(ctx, reqMessage)
	//	if err != nil {
	//		m.log.Error("error send message req: %v, err: %v", req, err)
	//	}
	//}()

	return reqMessage.Text, nil
}

func (m *MessageUseCase) processUniCheck(ctx context.Context, req entity.MessageRequest) {
	if _, valid := UniversityPreferences[req.FromNo]; !valid {
		up := m.getUniversityFromMessage(ctx, req.Data.Text)
		if len(up) <= 0 {
			msg := common.PrepareMessage(req, "Ok, no problem! We'll help you figure out what you're looking for 🥂. But before we get started, please help me answer one more question", "")
			m.ada.SendMessage(ctx, msg)
		}

		UniversityPreferences[req.FromNo] = up

		msg := common.PrepareMessage(req, "Do you have a *major* in mind that you're interested in?", "")
		m.ada.SendMessage(ctx, msg)
		return
	}

	mp := m.getMajorFromMessage(ctx, req.Data.Text)
	//if len(mp) > 0 {
	MajorPreferences[req.FromNo] = mp
	//}

	msg := common.PrepareMessageButton(req, "Mau apa kamu?", "header", "footer", []string{"uni alert", "uni buddy", "uni connect"})
	m.ada.SendButton(ctx, msg)
}

func (m *MessageUseCase) processUniAlert(ctx context.Context, req entity.MessageRequest, uniPreferences []string) {
	if len(uniPreferences) <= 0 {
		if State[req.FromNo] == constant.UNI_ALERT {
			reqMessage := common.PrepareMessage(req, "Your input is not valid", "")

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

	message := fmt.Sprintf("Sure thing! We've already set up reminders to keep you in the loop about important events related to your registration timeline at: \n%s", strings.Join(uniPreferences, "\n"))

	reqMessage := common.PrepareMessage(req, message, "")

	err := m.ada.SendMessage(ctx, reqMessage)
	if err != nil {
		m.log.Error("askUniversityPreferences Failed Send Message, err: %v", err)
	}
}

func (m *MessageUseCase) prosesUniConnect(ctx context.Context, req entity.MessageRequest) {

}

func (m *MessageUseCase) askUniversityPreferences(ctx context.Context, req entity.MessageRequest) error {
	reqMessage := common.PrepareMessage(req, "Do you have a university in mind that you're interested in?", "")
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

func (m *MessageUseCase) processUniBuddy(ctx context.Context, req entity.MessageRequest) *entity.AdaRequest {
	bisonChatReq := initBisonChatUniBuddyRequest()
	var messages []entity.Message
	if val, ok := UniBuddy[req.FromNo]; ok {
		messages = val
	}

	fmt.Printf("user: %s\ncurrent message uni-buddy: %v\n", req.FromNo, messages)

	messages = append(messages, entity.Message{
		Author:  "user",
		Content: req.Data.Text,
	})

	bisonChatReq.Instances[0].Messages = messages
	res, err := m.vertex.DoCallVertexAPIChat(ctx, bisonChatReq, m.getAccessToken())
	if err != nil {
		m.log.Error("DoCallVertexAPIChat error: %v\nmessage: %v", err, messages)
		return nil
	}

	newMessage := res.Predictions[0].Candidates[0]
	if !strings.EqualFold("bot", newMessage.Author) {
		newMessage.Author = "bot"
	}
	messages = append(messages, newMessage)
	UniBuddy[req.FromNo] = messages

	/* todo: check message:
	 	1. ended;
		2. check kalo {"linkUrl": http://aasdf.com, "type": "image", "message": "Message Sample"}
	*/

	return prepareMessageToUser(req, res.Predictions[0].Candidates[0].Content, "text")
}

func prepareMessageToUser(req entity.MessageRequest, message, messageType string) *entity.AdaRequest {
	return &entity.AdaRequest{
		Platform: req.Platform,
		From:     req.AccountNo,
		To:       req.Data.CustNo,
		Type:     messageType,
		Text:     message,
	}
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

func initBisonChatUniBuddyRequest() entity.BisonChatRequest {
	return entity.BisonChatRequest{
		Instances: []entity.Instance{
			{
				Context:  constant.ContextBisonChatUniBuddy,
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

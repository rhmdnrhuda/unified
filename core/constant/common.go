package constant

import (
	"github.com/rhmdnrhuda/unified/core/entity"
	"time"
)

const (
	RedisAccessTokenKey = "access-token-key"
	RedisAccessTokenTTL = 50 * time.Minute

	//	Feature
	FEATURE_UNI_CHECK   = "uni-check"
	FEATURE_UNI_BUDDY   = "uni-buddy"
	FEATURE_UNI_CONNECT = "uni-connect"
	FEATURE_UNI_ALERT   = "uni-alert"
)

var (
	TemplatePromptBisonChat        = "I am UNIFIED a chatbot that have feature uni-check for onboarding user like user said hi, hello, how are you, etc; uni-alert: can give alert or remind for student about registration to university; uni-buddy: help student to findout or give recommendation about which university and major that can fit with; uni-connect: help student to connect with university student for consultation 1on1 about university. Help to clasify this message to JSON format university_prefereces: [university_name], major_prefereces: [major], feature: unified_feature. when message from user is: %s"
	TemplatePromptGetUniTimeline   = "give me timeline of registration to institute teknolog bandung (ITB). format the response into array of struct with field title, year, month and date"
	TemplateValidateUniversityName = "please verify this message is contain valid university name or not, reply me with no or university name only. message: %s"
	TemplateValidateMajor          = "please verify this message is contain valid major or not, reply me major name (without other text and make Pascal Case with space) if not valid reply with no. message: %s"
	ContextBisonChatUniBuddy       = "I am Unified, a student personal assistant. I help students choose universities and majors by providing recommendations and answering their questions with a friendly manner.\nTo start the feature, show general description regarding Universitas Indonesia while also showing information of popular major on the Universitas Indonesia\nAlso ask user (choice 1) whether they have any question or (choice 2) if user wants us to guide them with some questions along with the information\nGive output in the form of JSON in the format of \n{\"linkUrl\": http://aasdf.com, \"type\": \"image\", \"message\": \"Message Sample\"}\n1. If the answer is already specific to a certain university and/or major, linkUrl will be a youtube link that contains campus profile related and also \"type\" is \"video\". But if there are no related results on Youtube, fill linkUrl with the image of the university and \"type\" is set as \"photo\". \n2. If the answer is not specified with the the university and/or major but the answer contains the recommendation, linkUrL will be a picture llink that contains recommendation campus/major  and  \"type\" is set as \"photo\"\n3. else give output in the JSON format:\n{\"message\": \"Message Sample\"}\nMessage will be referred to the related answer.\nIf user does not have any other questions or user already is interested to a specific university and major, send output JSON in this format:\n{\"university\": UNIVERSITY_NAME, \"major\": MAJOR_NAME}.\nuniversity refers to the university that user is specifying, if none then send \"university\" is NULL\nmajor refers to the major that user is specifying, if none then send \"major\" is NULL."
	ExampleBisonChatUniBuddy       = []entity.Example{
		//{
		//	Input: entity.Content{
		//		Content: "join uni-buddy",
		//	},
		//	Output: entity.Content{
		//		Content: "Do you have any university preference?",
		//	},
		//},
		{
			Input: entity.Content{
				Content: "hi",
			},
			Output: entity.Content{
				Content: "Hi, Do you have any major preference?",
			},
		},
	}
)

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Arman92/go-tdlib"
)

var client *tdlib.Client

func main() {
	tdlib.SetLogVerbosityLevel(1)
	tdlib.SetFilePath("./errors.txt")

	// Create new instance of client
	client = tdlib.NewClient(tdlib.Config{
		APIID:               "FILL YOUR API ID HERE",
		APIHash:             "FILL YOUR API HASH HERE",
		SystemLanguageCode:  "en",
		DeviceModel:         "Server",
		SystemVersion:       "1.0.0",
		ApplicationVersion:  "1.0.0",
		UseMessageDatabase:  true,
		UseFileDatabase:     true,
		UseChatInfoDatabase: true,
		UseTestDataCenter:   false,
		DatabaseDirectory:   "./tdlib-db",
		FileDirectory:       "./tdlib-files",
		IgnoreFileNames:     false,
	})

	// Handle Ctrl+C , Gracefully exit and shutdown tdlib
	var ch = make(chan os.Signal, 2)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		client.DestroyInstance()
		os.Exit(1)
	}()

	go func() {
		for {
			var currentState, _ = client.Authorize()
			if currentState.GetAuthorizationStateEnum() == tdlib.AuthorizationStateWaitPhoneNumberType {
				fmt.Print("Enter phone: ")
				var number string
				fmt.Scanln(&number)
				var _, err = client.SendPhoneNumber(number)
				if err != nil {
					fmt.Printf("Error sending phone number: %v\n", err)
				}
			} else if currentState.GetAuthorizationStateEnum() == tdlib.AuthorizationStateWaitCodeType {
				fmt.Print("Enter code: ")
				var code string
				fmt.Scanln(&code)
				var _, err = client.SendAuthCode(code)
				if err != nil {
					fmt.Printf("Error sending auth code : %v\n", err)
				}
			} else if currentState.GetAuthorizationStateEnum() == tdlib.AuthorizationStateWaitPasswordType {
				fmt.Print("Enter Password: ")
				var password string
				fmt.Scanln(&password)
				var _, err = client.SendAuthPassword(password)
				if err != nil {
					fmt.Printf("Error sending auth password: %v\n", err)
				}
			} else if currentState.GetAuthorizationStateEnum() == tdlib.AuthorizationStateReadyType {
				fmt.Println("Authorization Ready! Let's rock")
				break
			}
		}
	}()

	// Wait while we get Authorization Ready!
	// Note: See authorization example for complete auhtorization sequence example
	var currentState, _ = client.Authorize()
	for ; currentState.GetAuthorizationStateEnum() != tdlib.AuthorizationStateReadyType; currentState, _ = client.Authorize() {
		time.Sleep(300 * time.Millisecond)
	}

	http.HandleFunc("/getChats", getChatsHandler)
	http.ListenAndServe(":3000", nil)
}

func getChatsHandler(w http.ResponseWriter, req *http.Request) {
	var allChats, getChatErr = getChatList(client, 1000)
	if getChatErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(getChatErr.Error()))
		return
	}

	var retMap = make(map[string]interface{})
	retMap["total"] = len(allChats)

	var chatTitles []string
	for _, chat := range allChats {
		chatTitles = append(chatTitles, chat.Title)
	}

	retMap["chatList"] = chatTitles

	var ret, marshalErr = json.Marshal(retMap)
	if marshalErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(marshalErr.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, string(ret))
}

// see https://stackoverflow.com/questions/37782348/how-to-use-getchats-in-tdlib
func getChatList(client *tdlib.Client, limit int) ([]*tdlib.Chat, error) {
	var allChats []*tdlib.Chat
	var offsetOrder = int64(math.MaxInt64)
	var offsetChatID = int64(0)
	var chatList = tdlib.NewChatListMain()
	var lastChat *tdlib.Chat

	for len(allChats) < limit {
		if len(allChats) > 0 {
			lastChat = allChats[len(allChats)-1]
			for i := 0; i < len(lastChat.Positions); i++ {
				//Find the main chat list
				if lastChat.Positions[i].List.GetChatListEnum() == tdlib.ChatListMainType {
					offsetOrder = int64(lastChat.Positions[i].Order)
				}
			}
			offsetChatID = lastChat.ID
		}

		// get chats (ids) from tdlib
		var chats, getChatsErr = client.GetChats(chatList, tdlib.JSONInt64(offsetOrder),
			offsetChatID, int32(limit-len(allChats)))
		if getChatsErr != nil {
			return nil, getChatsErr
		}
		if len(chats.ChatIDs) == 0 {
			return allChats, nil
		}

		for _, chatID := range chats.ChatIDs {
			// get chat info from tdlib
			var chat, getChatErr = client.GetChat(chatID)
			if getChatErr == nil {
				allChats = append(allChats, chat)
			} else {
				return nil, getChatErr
			}
		}
	}

	return allChats, nil
}

package main

import (
	"encoding/json"
	"fmt"
	"github.com/mattermost/mattermost-plugin-starter-template/server/command"
	"github.com/mattermost/mattermost-plugin-starter-template/server/store/kvstore"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/mattermost/mattermost/server/public/pluginapi"
	"github.com/mattermost/mattermost/server/public/pluginapi/cluster"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Plugin struct {
	plugin.MattermostPlugin

	// kvstore is the client used to read/write KV records for this plugin.
	kvstore kvstore.KVStore

	// client is the Mattermost server API client.
	client *pluginapi.Client

	// commandClient is the client used to register and execute slash commands.
	commandClient command.Command

	backgroundJob *cluster.Job

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration
}
type MessageStatus map[string]string

const (
	// StatusDelivered означает, что сообщение доставлено, но не прочитано.
	StatusDelivered = "delivered"
	// StatusRead означает, что сообщение прочитано.
	StatusRead = "read"
)

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	parts := strings.Fields(args.Command)
	var postID string

	if len(parts) >= 2 {
		postID = parts[1]
	} else if args.RootId != "" {
		// Если команда вызвана как reply — пересылаем исходный пост треда
		postID = args.RootId
	} else {
		return &model.CommandResponse{
			Text:         "Используйте /forward [post_id] или вызовите команду как reply на нужное сообщение.",
			ResponseType: model.CommandResponseTypeEphemeral,
		}, nil
	}

	// Открываем интерактивный диалог (работает и в мобильных, и в вебе)
	dialog := model.OpenDialogRequest{
		TriggerId: args.TriggerId,
		URL:       fmt.Sprintf("/plugins/%s/submit", manifest.Id),
		Dialog: model.Dialog{
			CallbackId:       fmt.Sprintf("forward_message_%s", postID),
			Title:            "Переслать сообщение",
			IntroductionText: "Выберите получателя для пересылки сообщения",
			Elements: []model.DialogElement{
				{
					DisplayName: "Пользователь",
					Name:        "user_recipient",
					Type:        "select",
					Placeholder: "Выберите пользователя...",
					DataSource:  "users",
					Optional:    true,
				},
				{
					DisplayName: "Канал",
					Name:        "channel_recipient",
					Type:        "select",
					Placeholder: "Выберите канал...",
					DataSource:  "channels",
					Optional:    true,
				},
			},
			SubmitLabel:    "Переслать",
			NotifyOnCancel: false,
		},
	}

	if err := p.API.OpenInteractiveDialog(dialog); err != nil {
		return &model.CommandResponse{
			Text:         "Ошибка открытия диалога: " + err.Error(),
			ResponseType: model.CommandResponseTypeEphemeral,
		}, nil
	}

	return &model.CommandResponse{}, nil
}

func (p *Plugin) handleDialogSubmit(w http.ResponseWriter, r *http.Request) {
	var submission model.SubmitDialogRequest
	if err := json.NewDecoder(r.Body).Decode(&submission); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Извлекаем post_id из callback_id
	callbackParts := strings.Split(submission.CallbackId, "_")
	if len(callbackParts) < 3 {
		http.Error(w, "Invalid callback_id", http.StatusBadRequest)
		return
	}
	postID := callbackParts[2]

	userRecipient := submission.Submission["user_recipient"]
	channelRecipient := submission.Submission["channel_recipient"]

	var targetChannelID string
	if userRecipient != nil && userRecipient.(string) != "" {
		directChannel, err := p.API.GetDirectChannel(submission.UserId, userRecipient.(string))
		if err != nil {
			http.Error(w, "Ошибка создания личного канала: "+err.Error(), http.StatusInternalServerError)
			return
		}
		targetChannelID = directChannel.Id
	} else if channelRecipient != nil && channelRecipient.(string) != "" {
		targetChannelID = channelRecipient.(string)
	} else {
		http.Error(w, "Необходимо выбрать получателя", http.StatusBadRequest)
		return
	}

	if err := p.forwardMessage(postID, targetChannelID, submission.UserId); err != nil {
		http.Error(w, "Ошибка пересылки: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (p *Plugin) forwardMessage(postID, targetChannelID, senderID string) error {
	// Получаем оригинальный пост
	post, appErr := p.API.GetPost(postID)
	if appErr != nil {
		return appErr
	}

	// Получаем пользователя-автора оригинального поста
	originalUser, appErr := p.API.GetUser(post.UserId)
	if appErr != nil {
		return appErr
	}

	t := time.Unix(post.CreateAt/1000, 0)
	timeStr := t.Format("02.01.2006 15:04")

	// Формируем текст пересланного сообщения
	forwardedText := fmt.Sprintf(
		"📨 **Переслано сообщение**\n"+
			"**Автор:** %s (@%s)\n"+
			"**Время:** %s\n\n"+
			"%s",
		originalUser.GetDisplayName(model.ShowNicknameFullName),
		originalUser.Username,
		timeStr,
		post.Message,
	)

	// Создаем новый пост с файлами
	newPost := &model.Post{
		ChannelId: targetChannelID,
		UserId:    senderID,
		Message:   forwardedText,
		FileIds:   post.FileIds,
		Type:      "custom_forwarded",
	}
	_, appErr = p.API.CreatePost(newPost)
	return appErr
}
func (p *Plugin) OnActivate() error {
	return p.API.RegisterCommand(&model.Command{
		Trigger:          "forward",
		DisplayName:      "Forward Message",
		Description:      "Переслать сообщение другому пользователю или в канал",
		AutoComplete:     true,
		AutoCompleteDesc: "/forward [post_id]",
		AutoCompleteHint: "[post_id]",
	})
}

func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/forward":
		p.handleForward(w, r)
	case "/submit":
		p.handleDialogSubmit(w, r)
	case "/recipients_list":
		p.handleRecipientsList(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (p *Plugin) handleForward(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("Mattermost-User-Id")
	if userID == "" {
		http.Error(w, "not authorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		PostID    string `json:"post_id"`
		Recipient string `json:"recipient"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	var channelID string
	if strings.HasPrefix(req.Recipient, "user:") {
		targetUserID := strings.TrimPrefix(req.Recipient, "user:")
		ch, appErr := p.API.GetDirectChannel(userID, targetUserID)
		if appErr != nil {
			http.Error(w, appErr.Error(), http.StatusInternalServerError)
			return
		}
		channelID = ch.Id
	} else if strings.HasPrefix(req.Recipient, "channel:") {
		channelID = strings.TrimPrefix(req.Recipient, "channel:")
	} else {
		http.Error(w, "invalid recipient", http.StatusBadRequest)
		return
	}

	post, appErr := p.API.GetPost(req.PostID)
	if appErr != nil {
		http.Error(w, appErr.Error(), http.StatusInternalServerError)
		return
	}

	originalUser, appErr := p.API.GetUser(post.UserId)
	if appErr != nil {
		http.Error(w, appErr.Error(), http.StatusInternalServerError)
		return
	}

	newPost := &model.Post{
		ChannelId: channelID,
		UserId:    userID,
		Message:   post.Message,
		FileIds:   post.FileIds,
		Type:      "custom_forwarded",
		Props: map[string]interface{}{
			"original_display_name": originalUser.GetDisplayName(model.ShowNicknameFullName),
			"original_username":     originalUser.Username,
			"original_create_at":    post.CreateAt,
		},
	}
	_, appErr = p.API.CreatePost(newPost)
	if appErr != nil {
		http.Error(w, appErr.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (p *Plugin) getRecipientOptions(userID string) []*model.PostActionOptions {
	// Получение пользователей и каналов, формирование options
	// Пример для пользователей:
	options := []*model.PostActionOptions{}
	users, _ := p.API.GetUsers(&model.UserGetOptions{Page: 0, PerPage: 50})
	for _, user := range users {
		if user.Id != userID && !user.IsBot {
			options = append(options, &model.PostActionOptions{
				Text:  fmt.Sprintf("👤 %s (@%s)", user.GetDisplayName(model.ShowNicknameFullName), user.Username),
				Value: "user:" + user.Id,
			})
		}
	}
	// Аналогично добавьте каналы
	return options
}

func (p *Plugin) handleRecipientsList(w http.ResponseWriter, r *http.Request) {
	// Получаем первых 100 пользователей
	users, appErr := p.API.GetUsers(&model.UserGetOptions{
		Page:    0,
		PerPage: 100, // увеличьте лимит при необходимости
	})
	if appErr != nil {
		http.Error(w, appErr.Error(), http.StatusInternalServerError)
		return
	}

	// Получаем первую команду и её каналы
	teams, appErr := p.API.GetTeams()
	if appErr != nil || len(teams) == 0 {
		http.Error(w, "No teams found", http.StatusInternalServerError)
		return
	}
	teamID := teams[1].Id

	channels, appErr := p.API.GetPublicChannelsForTeam(teamID, 0, 100)
	if appErr != nil {
		http.Error(w, appErr.Error(), http.StatusInternalServerError)
		return
	}

	var result []map[string]string
	for _, u := range users {
		result = append(result, map[string]string{
			"type":  "user",
			"id":    u.Id,
			"label": "👤 " + u.GetDisplayName(model.ShowFullName) + " (@" + u.Username + ")",
		})
	}
	for _, c := range channels {
		result = append(result, map[string]string{
			"type":  "channel",
			"id":    c.Id,
			"label": "📢 " + c.DisplayName,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

// MessageHasBeenPosted is invoked after a message has been posted to a channel.
// Мы будем использовать этот хук для записи начального статуса "delivered" для каждого получателя.
func (p *Plugin) MessageHasBeenPosted(c *plugin.Context, post *model.Post) {
	// 1. Логгирование для проверки, что хук сработал
	p.API.LogDebug("Hook MessageHasBeenPosted triggered", "postID", post.Id, "channelID", post.ChannelId)

	// Игнорируем сообщения от ботов или системные сообщения
	if post.UserId == "" || post.IsSystemMessage() {
		p.API.LogDebug("Skipping post because it is from a bot or system", "postID", post.Id)
		return
	}

	// 2. Получаем всех участников канала
	// Примечание: для больших каналов может потребоваться постраничная загрузка.
	// Для примера мы загружаем до 200 участников.
	channelMembers, err := p.API.GetChannelMembers(post.ChannelId, 0, 200)
	if err != nil {
		p.API.LogError("Failed to get channel members", "channelID", post.ChannelId, "error", err)
		return
	}

	// 3. Создаем карту статусов
	statuses := make(MessageStatus)
	for _, member := range channelMembers {
		// Статус не нужен для самого отправителя
		if member.UserId == post.UserId {
			continue
		}
		statuses[member.UserId] = StatusDelivered
	}

	// Если в канале нет других участников, ничего не делаем
	if len(statuses) == 0 {
		p.API.LogDebug("No other recipients in the channel, skipping status save.", "postID", post.Id)
		return
	}

	p.API.LogDebug("Initialized statuses for post", "postID", post.Id, "statuses", statuses)

	// 4. Сериализуем карту в JSON для сохранения в KVStore
	jsonValue, jsonErr := json.Marshal(statuses)
	if jsonErr != nil {
		p.API.LogError("Failed to marshal message statuses to JSON", "postID", post.Id, "error", jsonErr)
		return
	}

	// 5. Сохраняем в KVStore
	// Ключ будет уникальным для каждого сообщения
	key := "status_" + post.Id
	// Используем p.API.KVSet для сохранения данных. Это стандартный способ работы с KVStore [2].
	if appErr := p.API.KVSet(key, jsonValue); appErr != nil {
		p.API.LogError("Failed to save message statuses to KVStore", "postID", post.Id, "key", key, "error", appErr)
		return
	}

	// Финальное логгирование успеха
	// p.API.LogInfo, p.API.LogDebug, p.API.LogError - стандартные методы для логгирования в плагинах [3][4][5].
	p.API.LogInfo("Successfully stored message statuses", "postID", post.Id, "key", key)
}

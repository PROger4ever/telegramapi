package telegramapi

import (
	"log"
	"time"

	"github.com/PROger4ever/telegramapi/mtproto"
)

func (c *Conn) LoadChats(contacts *ContactList, limit int, offsetPeer mtproto.TLInputPeerType) error {
	log.Printf("Loading list of chats...")
	r, err := c.Send(&mtproto.TLMessagesGetDialogs{
		Flags:      0,
		Limit:      limit,
		OffsetPeer: offsetPeer,
	})
	if err != nil {
		return err
	}
	switch r := r.(type) {
	case *mtproto.TLMessagesDialogs:
		c.updateChatsLocked(contacts, r.Dialogs, r.Messages, r.Chats, r.Users)
		return nil
	case *mtproto.TLMessagesDialogsSlice:
		c.updateChatsLocked(contacts, r.Dialogs, r.Messages, r.Chats, r.Users)
		return nil
	default:
		return c.HandleUnknownReply(r)
	}
}

func (c *Conn) updateChatsLocked(contacts *ContactList, dialogs []*mtproto.TLDialog, apimessages []mtproto.TLMessageType, chats []mtproto.TLChatType, users []mtproto.TLUserType) {
	c.stateMut.Lock()
	defer c.stateMut.Unlock()

	accessHashByUserID := make(map[int]uint64)
	accessHashByChannelID := make(map[int]uint64)
	c.updateUsers(contacts, users, accessHashByUserID)
	c.updateGroups(contacts, chats, accessHashByChannelID)

	// for _, apimsg := range apimessages {
	// 	msg := c.updateMessage(contacts, messages, apimsg)
	// 	if msg != nil {
	// 		msgs = append(msgs, msg)
	// 	}
	// }

	contacts.Chats = contacts.Chats[:0]

	for _, dialog := range dialogs {
		var chat *Chat
		if upeer, ok := dialog.Peer.(*mtproto.TLPeerUser); ok {
			if user := contacts.Users[upeer.UserID]; user != nil {
				chat = contacts.UserChats[user.ID]
				if chat == nil {
					chat = &Chat{
						Type:     UserChat,
						ID:       user.ID,
						User:     user,
						Messages: newMessageList(),
					}
					contacts.UserChats[user.ID] = chat
				}

				chat.AccessHash = accessHashByUserID[user.ID]
				chat.Username = user.Username

				if user == contacts.Self {
					contacts.SelfChat = chat
				}

				// TODO: dialog.TopMessage
			}
		} else if cpeer, ok := dialog.Peer.(*mtproto.TLPeerChannel); ok {
			chat = contacts.ChannelChats[cpeer.ChannelID]
			if chat == nil {
				chat = &Chat{
					Type:     ChannelChat,
					ID:       cpeer.ChannelID,
					Messages: newMessageList(),
				}
				contacts.ChannelChats[cpeer.ChannelID] = chat
			}
			chat.AccessHash = accessHashByChannelID[cpeer.ChannelID]
			if channel := contacts.Channels[cpeer.ChannelID]; channel != nil {
				chat.Title = channel.Title
			}
		} else if gpeer, ok := dialog.Peer.(*mtproto.TLPeerChat); ok {
			if group := contacts.Groups[gpeer.ChatID]; group != nil {
				chat = contacts.GroupChats[group.ID]
				if chat == nil {
					chat = &Chat{
						Type:     GroupChat,
						ID:       group.ID,
						Messages: newMessageList(),
					}
					contacts.GroupChats[group.ID] = chat
				}
				chat.Title = group.Title
			}
		} else {
			log.Printf("Unknown dialog peer: %v", dialog)
		}
		if chat != nil {
			contacts.Chats = append(contacts.Chats, chat)
		}
	}
}

func (c *Conn) LoadHistory(contacts *ContactList, chat *Chat, limit int) error {
	more := true
	var count int
	log.Printf("Loading history of “%s”...", chat.TitleOrName())
	for more && (limit == 0 || count < limit) {
		r, err := c.Send(&mtproto.TLMessagesGetHistory{
			Peer:     chat.inputPeer(),
			Limit:    10000,
			OffsetID: chat.Messages.MinKnownID,
		})
		if err != nil {
			return err
		}
		switch r := r.(type) {
		case *mtproto.TLMessagesMessages:
			c.updateHistoryLocked(contacts, chat, r.Messages, r.Chats, r.Users)
			more = false
			count += len(r.Messages)
		case *mtproto.TLMessagesMessagesSlice:
			c.updateHistoryLocked(contacts, chat, r.Messages, r.Chats, r.Users)
			more = len(r.Messages) > 0
			count += len(r.Messages)
		case *mtproto.TLMessagesChannelMessages:
			c.updateHistoryLocked(contacts, chat, r.Messages, r.Chats, r.Users)
			more = len(r.Messages) > 0
			count += len(r.Messages)
		default:
			return c.HandleUnknownReply(r)
		}
		if more {
			log.Printf("Loaded %d messages...", count)
		}

		time.Sleep(1 * time.Second)
	}
	log.Printf("Done. Loaded %d messages.", count)

	return nil
}

func (c *Conn) updateHistoryLocked(contacts *ContactList, chat *Chat, apimessages []mtproto.TLMessageType, chats []mtproto.TLChatType, users []mtproto.TLUserType) {
	c.stateMut.Lock()
	defer c.stateMut.Unlock()

	c.updateUsers(contacts, users, nil)

	messages := chat.Messages

	var msgs []*Message
	for _, apimsg := range apimessages {
		msg := c.updateMessage(contacts, messages, apimsg)
		if msg != nil {
			msgs = append(msgs, msg)
		}
	}

	for i, j := 0, len(msgs)-1; i < j; {
		msgs[i], msgs[j] = msgs[j], msgs[i]
		i++
		j--
	}

	messages.Messages = append(msgs, messages.Messages...)
}

func (c *Conn) updateMessage(contacts *ContactList, messages *MessageList, apimsg mtproto.TLMessageType) *Message {
	switch apimsg := apimsg.(type) {
	case *mtproto.TLMessage:
		messages.foundID(apimsg.ID)

		if apimsg.Message == "" {
			return nil
		}

		var fromu *User
		if apimsg.FromID != 0 {
			fromu = contacts.Users[apimsg.FromID]
			if fromu == nil {
				return nil
			}
		}

		msg := messages.MessagesByID[apimsg.ID]
		if msg == nil {
			msg = &Message{
				ID: apimsg.ID,
			}
			messages.MessagesByID[apimsg.ID] = msg
		}

		msg.Type = NormalMessage

		msg.Date = makeDate(apimsg.Date)
		msg.EditDate = makeDate(apimsg.EditDate)

		msg.From = fromu

		msg.ReplyToID = apimsg.ReplyToMsgID

		msg.Text = apimsg.Message

		if apimsg.FwdFrom != nil {
			msg.FwdFrom = contacts.Users[apimsg.FwdFrom.FromID]
			msg.FwdDate = makeDate(apimsg.FwdFrom.Date)
		}

		return msg

	case *mtproto.TLMessageService:
		messages.foundID(apimsg.ID)
		return nil

	case *mtproto.TLMessageEmpty:
		messages.foundID(apimsg.ID)
		return nil

	default:
		return nil
	}
}

func (c *Conn) updateUsers(contacts *ContactList, users []mtproto.TLUserType, accessHashByUserID map[int]uint64) {
	selfUserID := c.state.UserID
	for _, apiuser := range users {
		if apiuser, ok := apiuser.(*mtproto.TLUser); ok {
			user := contacts.Users[apiuser.ID]
			if user == nil {
				user = &User{ID: apiuser.ID}
				contacts.Users[apiuser.ID] = user
			}
			user.Username = apiuser.Username
			user.FirstName = apiuser.FirstName
			user.LastName = apiuser.LastName
			if accessHashByUserID != nil {
				accessHashByUserID[apiuser.ID] = apiuser.AccessHash
			}
			if user.ID == selfUserID {
				contacts.Self = user
			}
		}
	}
}

func (c *Conn) updateGroups(contacts *ContactList, apichats []mtproto.TLChatType, accessHashByChannelID map[int]uint64) {
	for _, apichat := range apichats {
		if apigroup, ok := apichat.(*mtproto.TLChat); ok {
			group := contacts.Groups[apigroup.ID]
			if group == nil {
				group = &Group{ID: apigroup.ID}
				contacts.Groups[apigroup.ID] = group
			}
			group.Title = apigroup.Title
			group.ParticipantsCount = apigroup.ParticipantsCount
		} else if apichan, ok := apichat.(*mtproto.TLChannel); ok {
			channel := contacts.Channels[apichan.ID]
			if channel == nil {
				channel = &Channel{ID: apichan.ID}
				contacts.Channels[apichan.ID] = channel
			}
			channel.Title = apichan.Title
			if accessHashByChannelID != nil {
				accessHashByChannelID[apichan.ID] = apichan.AccessHash
			}
		}
	}
}

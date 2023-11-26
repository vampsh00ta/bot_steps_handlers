package bot

import (
	"regexp"
	"strings"

	"github.com/go-telegram/bot/models"
)

type HandlerType int

const (
	HandlerTypeMessageText HandlerType = iota
	HandlerTypeCallbackQueryData
)

type MatchType int

const (
	MatchTypeExact MatchType = iota
	MatchTypePrefix
	MatchTypeContains

	matchTypeRegexp
	matchTypeFunc
)

type handler struct {
	handlerType HandlerType
	matchType   MatchType
	handler     HandlerFunc
	userId      int64
	pattern     string
	fmt         interface{}
	re          *regexp.Regexp
	matchFunc   MatchFunc
}

func (h handler) match(update *models.Update) bool {
	if h.matchType == matchTypeFunc {
		return h.matchFunc(update)
	}
	var data string
	switch h.handlerType {
	case HandlerTypeMessageText:
		data = update.Message.Text
	case HandlerTypeCallbackQueryData:
		data = update.CallbackQuery.Data
	}
	if h.userId != 0 && update.Message.From.ID != h.userId {
		return false
	}

	if h.matchType == MatchTypeExact {
		return data == h.pattern
	}
	if h.matchType == MatchTypePrefix {
		return strings.HasPrefix(data, h.pattern)
	}
	if h.matchType == MatchTypeContains {
		return strings.Contains(data, h.pattern)
	}
	if h.matchType == matchTypeRegexp {
		return h.re.Match([]byte(data))
	}
	return false
}

func (b *Bot) RegisterHandlerMatchFunc(matchFunc MatchFunc, f HandlerFunc) string {
	b.handlersMx.Lock()
	defer b.handlersMx.Unlock()

	id := RandomString(16)

	h := handler{
		matchType: matchTypeFunc,
		matchFunc: matchFunc,
		handler:   f,
	}

	b.handlers[id] = h

	return id
}

func (b *Bot) RegisterHandlerRegexp(handlerType HandlerType, re *regexp.Regexp, f HandlerFunc) string {
	b.handlersMx.Lock()
	defer b.handlersMx.Unlock()

	id := RandomString(16)

	h := handler{
		handlerType: handlerType,
		matchType:   matchTypeRegexp,
		re:          re,
		handler:     f,
	}

	b.handlers[id] = h

	return id
}

func (b *Bot) RegisterHandler(handlerType HandlerType, pattern string, matchType MatchType, f HandlerFunc) string {
	b.handlersMx.Lock()
	defer b.handlersMx.Unlock()

	id := RandomString(16)

	h := handler{
		handlerType: handlerType,
		matchType:   matchType,
		pattern:     pattern,
		handler:     f,
	}

	b.handlers[id] = h

	return id
}

func (b *Bot) UnregisterHandler(id string) {
	b.handlersMx.Lock()
	defer b.handlersMx.Unlock()

	delete(b.handlers, id)
}

func (b *Bot) RegisterStepHandler(userId int64, nextFunc HandlerFunc, data interface{}) string {
	b.UnregisterStepHandler(userId)
	b.handlersMx.Lock()
	defer b.handlersMx.Unlock()

	id := RandomString(16)

	h := handler{
		handlerType: HandlerTypeMessageText,
		matchType:   MatchTypeContains,
		pattern:     "",
		handler:     nextFunc,
		userId:      userId,
		fmt:         data,
	}

	b.handlers[id] = h
	b.stepHandlers[userId] = id
	b.stepHanderData[userId] = data

	return id

}

func (b *Bot) GetStepData(userId int64) interface{} {
	b.stepMx.Lock()
	defer b.stepMx.Unlock()
	data, ok := b.stepHanderData[userId]
	if !ok {
		return nil
	}
	return data
}
func (b *Bot) UnregisterStepHandler(userId int64) {
	b.handlersMx.Lock()
	defer b.handlersMx.Unlock()

	id := b.stepHandlers[userId]

	delete(b.handlers, id)
	delete(b.stepHandlers, userId)

}

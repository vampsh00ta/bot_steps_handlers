package bot

import (
	"context"
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

	pattern   string
	re        *regexp.Regexp
	matchFunc MatchFunc
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

func (bot *Bot) RegisterStepHandler(ctx context.Context, update *models.Update, nextFunc HandlerFunc, data any) {
	bot.stepMx.RLock()
	me, _ := bot.GetMe(ctx)
	stepId, ok := bot.stepHandlerId[me.ID]
	bot.stepMx.RUnlock()

	if ok {
		bot.UnregisterHandler(stepId)

	}
	bot.stepMx.Lock()
	defer bot.stepMx.Unlock()
	stepId = bot.RegisterHandler(HandlerTypeMessageText, "", MatchTypeContains, nextFunc)
	bot.stepHanderData[stepId] = data
	bot.stepHandlerId[me.ID] = stepId

}
func (bot *Bot) UnregisterStepHandler(ctx context.Context, update *models.Update) interface{} {
	me, _ := bot.GetMe(ctx)
	stepId, ok := bot.stepHandlerId[me.ID]
	if !ok {
		return nil

	}
	bot.handlersMx.Lock()
	data := bot.stepHanderData[stepId]
	delete(bot.stepHanderData, stepId)
	delete(bot.stepHandlerId, me.ID)
	bot.handlersMx.Unlock()
	bot.UnregisterHandler(stepId)
	return data
}
func (bot *Bot) GetStepData(ctx context.Context, update *models.Update) interface{} {
	me, _ := bot.GetMe(ctx)
	stepId, ok := bot.stepHandlerId[me.ID]
	if !ok {
		return nil

	}
	return bot.stepHanderData[stepId]
}

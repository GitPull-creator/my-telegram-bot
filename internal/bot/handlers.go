package bot

import (
	"database/sql"
	"log"
	"my-telegram-bot/internal/database"
	"my-telegram-bot/internal/storage"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func createMainKeyboard(db *sql.DB, userID int64) tgbotapi.InlineKeyboardMarkup {
	categories, err := storage.GetUserCategories(db, userID)
	if err != nil {
		log.Println("DB error:", err)
	}

	log.Printf("Получено %d категорий для пользователя %d", len(categories), userID)

	buttons := make([][]tgbotapi.InlineKeyboardButton, len(categories))
	for i, category := range categories {
		callbackData := "category:" + strconv.Itoa(category.ID)
		log.Printf("Создание кнопки: Name=%s, ID=%d, CallbackData=%s", category.Name, category.ID, callbackData)
		buttons[i] = []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(category.Name, callbackData),
		}
	}

	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

func handleStart(b *Bot, chatID int64, userID int64) {
	log.Printf("handleStart вызван для пользователя %d", userID)

	if err := database.CreateUserCategories(b.DB, userID); err != nil {
		log.Printf("Ошибка создания категорий для пользователя %d: %v", userID, err)
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка инициализации категорий. Попробуйте еще раз.")
		b.bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(chatID, "👋 Привет! Добро пожаловать в твою птичью галерею красоты. Выбери категорию:")
	msg.ReplyMarkup = createMainKeyboard(b.DB, userID)

	if _, err := b.bot.Send(msg); err != nil {
		log.Println("Ошибка отправки сообщения:", err)
	}
}

func handleUpdate(b *Bot, update tgbotapi.Update) {
	if update.Message != nil {
		if update.Message.Photo != nil {
			handlePhoto(b, update.Message)
			return
		}

		if update.Message.Text == "/start" {
			handleStart(b, update.Message.Chat.ID, update.Message.From.ID)
			return
		}

		if update.Message.Text != "" {
			handleText(b, update.Message)
		}
	}

	if update.CallbackQuery != nil {
		handleCallback(b, update.CallbackQuery)
	}
}

func handleCallback(b *Bot, callback *tgbotapi.CallbackQuery) {
	callbackData := callback.Data

	switch {
	case callbackData == "back_main":
		handleStart(b, callback.Message.Chat.ID, callback.From.ID)
	case len(callbackData) > 9 && callbackData[:9] == "category:":
		categoryID := callbackData[9:]
		handleCategorySelect(b, callback.Message.Chat.ID, callback.From.ID, categoryID)

	case len(callbackData) > 9 && callbackData[:9] == "add_card:":
		categoryID := callbackData[9:]
		handleAddCard(b, callback.Message.Chat.ID, callback.From.ID, categoryID)

	case len(callbackData) > 11 && callbackData[:11] == "show_cards:":
		categoryID := callbackData[11:]
		handleShowCards(b, callback.Message.Chat.ID, callback.From.ID, categoryID)

	case len(callbackData) > 16 && callbackData[:16] == "add_subcategory:":
		categoryID := callbackData[16:]
		handleAddSubcategory(b, callback.Message.Chat.ID, callback.From.ID, categoryID)

	default:
		log.Printf("Неизвестный callback: %s", callbackData)
		msg := tgbotapi.NewMessage(callback.Message.Chat.ID, "❌ Неверная команда")
		b.bot.Send(msg)
	}

	callback_answer := tgbotapi.NewCallback(callback.ID, "")
	b.bot.Request(callback_answer)
}

func handleCategorySelect(b *Bot, chatID int64, userID int64, categoryID string) {
	categoryIDInt, err := strconv.Atoi(categoryID)
	if err != nil {
		log.Printf("Ошибка преобразования categoryID: %s, ошибка: %v", categoryID, err)
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка идентификатора категории")
		b.bot.Send(msg)
		return
	}

	category, err := storage.GetCategoryByID(b.DB, userID, categoryIDInt)
	if err != nil {
		log.Println("DB error:", err)
		msg := tgbotapi.NewMessage(chatID, "❌ Категория не найдена")
		b.bot.Send(msg)
		return
	}

	if category.Name == "Косметика" {
		subcategories, err := storage.GetSubcategories(b.DB, userID, categoryIDInt)
		if err == nil && len(subcategories) > 0 {
			msg := tgbotapi.NewMessage(chatID, "Категория: "+category.Name+"\n\n📂 Подкатегории:")
			for _, sub := range subcategories {
				msg.Text += "\n• " + sub.Name
			}
			b.bot.Send(msg)
		}
	}

	msg := tgbotapi.NewMessage(chatID, "Вы выбрали категорию: "+category.Name)
	msg.ReplyMarkup = createCategoryKeyboard(categoryID, category.Name)

	if _, err := b.bot.Send(msg); err != nil {
		log.Println("Ошибка отправки сообщения:", err)
	}
}

func createCategoryKeyboard(categoryID string, categoryName string) tgbotapi.InlineKeyboardMarkup {
	buttons := [][]tgbotapi.InlineKeyboardButton{
		{
			tgbotapi.NewInlineKeyboardButtonData("➕ Добавить карточку", "add_card:"+categoryID),
			tgbotapi.NewInlineKeyboardButtonData("📋 Показать все", "show_cards:"+categoryID),
		},
	}

	if categoryName == "Косметика" {
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData("🔧 Добавить подкатегорию", "add_subcategory:"+categoryID),
		})
	}

	buttons = append(buttons, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", "back_main"),
	})

	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

func handleAddCard(b *Bot, chatID int64, userID int64, categoryID string) {
	categoryIDInt, err := strconv.Atoi(categoryID)
	if err != nil {
		log.Printf("Ошибка преобразования categoryID в handleAddCard: %s, ошибка: %v", categoryID, err)
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка идентификатора категории")
		b.bot.Send(msg)
		return
	}

	SetUserState(userID, "waiting_photo", categoryIDInt)
	msg := tgbotapi.NewMessage(chatID, "📸 Пришлите фото для карточки:")

	if _, err := b.bot.Send(msg); err != nil {
		log.Println("Ошибка отправки сообщения:", err)
	}
}

func handleShowCards(b *Bot, chatID int64, userID int64, categoryID string) {
	categoryIDInt, err := strconv.Atoi(categoryID)
	if err != nil {
		log.Printf("Ошибка преобразования categoryID: %s, ошибка: %v", categoryID, err)
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка идентификатора категории")
		b.bot.Send(msg)
		return
	}

	cards, err := storage.GetCategoryCards(b.DB, userID, categoryIDInt)
	if err != nil {
		log.Println("Ошибка получения карточек:", err)
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка при получении карточек")
		b.bot.Send(msg)
		return
	}

	if len(cards) == 0 {
		msg := tgbotapi.NewMessage(chatID, "📭 В этой категории пока нет карточек")
		b.bot.Send(msg)
	} else {
		for _, card := range cards {
			photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileID(card.PhotoFileID))

			caption := ""
			if card.Title != "" {
				caption += "📝 " + card.Title + "\n"
			}
			if card.Link != "" {
				caption += "🔗 " + card.Link
			}

			photo.Caption = caption
			b.bot.Send(photo)
		}
	}

	category, err := storage.GetCategoryByID(b.DB, userID, categoryIDInt)
	if err == nil {
		msg := tgbotapi.NewMessage(chatID, "Вы можете:")
		msg.ReplyMarkup = createCategoryKeyboard(categoryID, category.Name)
		b.bot.Send(msg)
	}
}

func handlePhoto(b *Bot, message *tgbotapi.Message) {
	userID := message.From.ID

	state, exists := GetUserState(userID)
	if !exists || state.State != "waiting_photo" {
		return
	}

	photo := message.Photo[len(message.Photo)-1]

	UpdateUserState(userID, UserState{
		PhotoFileID: photo.FileID,
		State:       "waiting_title",
	})

	msg := tgbotapi.NewMessage(message.Chat.ID, "✍️ Введите название карточки (или напишите 'пропустить'):")
	b.bot.Send(msg)
}

func handleText(b *Bot, message *tgbotapi.Message) {
	userID := message.From.ID
	text := message.Text

	state, exists := GetUserState(userID)
	if !exists {
		return
	}

	switch state.State {
	case "waiting_title":
		if text == "пропустить" {
			text = ""
		}

		UpdateUserState(userID, UserState{
			Title: text,
			State: "waiting_link",
		})

		msg := tgbotapi.NewMessage(message.Chat.ID, "🔗 Введите ссылку (или напишите 'пропустить'):")
		b.bot.Send(msg)

	case "waiting_link":
		link := text
		if text == "пропустить" {
			link = ""
		}

		saveCard(b, userID, message.Chat.ID, state, link)

		ClearUserState(userID)

	case "waiting_subcategory_name":
		subcategory := &database.Subcategory{
			Name:       text,
			CategoryID: state.CategoryID,
			UserID:     userID,
		}

		err := storage.AddSubcategory(b.DB, subcategory)
		if err != nil {
			log.Println("Ошибка сохранения подкатегории:", err)
			msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка при создании подкатегории")
			b.bot.Send(msg)
		} else {
			msg := tgbotapi.NewMessage(message.Chat.ID, "✅ Подкатегория '"+text+"' создана!")
			b.bot.Send(msg)

			msg2 := tgbotapi.NewMessage(message.Chat.ID, "👋 Выберите категорию:")
			msg2.ReplyMarkup = createMainKeyboard(b.DB, userID)
			b.bot.Send(msg2)
		}

		ClearUserState(userID)
	}
}

func saveCard(b *Bot, userID int64, chatID int64, state UserState, link string) {
	card := &database.Card{
		PhotoFileID:   state.PhotoFileID,
		Title:         state.Title,
		Link:          link,
		CategoryID:    state.CategoryID,
		SubcategoryID: nil,
		UserID:        userID,
	}

	err := storage.AddCard(b.DB, card)
	if err != nil {
		log.Println("Ошибка сохранения карточки:", err)
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка при сохранении карточки")
		b.bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(chatID, "✅ Карточка успешно добавлена!")
	b.bot.Send(msg)

	msg2 := tgbotapi.NewMessage(chatID, "👋 Выберите категорию:")
	msg2.ReplyMarkup = createMainKeyboard(b.DB, userID)
	b.bot.Send(msg2)
}

func handleAddSubcategory(b *Bot, chatID int64, userID int64, categoryID string) {
	categoryIDInt, err := strconv.Atoi(categoryID)
	if err != nil {
		log.Printf("Ошибка преобразования categoryID в handleAddSubcategory: %s, ошибка: %v", categoryID, err)
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка идентификатора категории")
		b.bot.Send(msg)
		return
	}

	SetUserState(userID, "waiting_subcategory_name", categoryIDInt)
	msg := tgbotapi.NewMessage(chatID, "✍️ Введите название подкатегории:")
	b.bot.Send(msg)
}

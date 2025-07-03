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
		log.Printf("Ошибка получения категорий: %v", err)
		return tgbotapi.NewInlineKeyboardMarkup()
	}

	if len(categories) == 0 {
		log.Printf("Нет категорий для пользователя %d", userID)
		// Возвращаем пустую клавиатуру
		return tgbotapi.NewInlineKeyboardMarkup()
	}

	buttons := make([][]tgbotapi.InlineKeyboardButton, len(categories))
	for i, category := range categories {
		buttons[i] = []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(category.Name, "category:"+strconv.Itoa(category.ID)),
		}
	}

	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

func handleStart(b *Bot, chatID int64, userID int64) {
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

		if update.Message.Text == "/reset" {
			handleReset(b, update.Message.Chat.ID, update.Message.From.ID)
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

	case len(callbackData) > 13 && callbackData[:13] == "subcategory:":
		subcategoryID := callbackData[13:]
		handleSubcategorySelect(b, callback.Message.Chat.ID, callback.From.ID, subcategoryID)

	case len(callbackData) > 17 && callbackData[:17] == "add_sub_card:":
		subcategoryID := callbackData[17:]
		handleAddSubcategoryCard(b, callback.Message.Chat.ID, callback.From.ID, subcategoryID)

	case len(callbackData) > 19 && callbackData[:19] == "show_sub_cards:":
		subcategoryID := callbackData[19:]
		handleShowSubcategoryCards(b, callback.Message.Chat.ID, callback.From.ID, subcategoryID)

	case len(callbackData) > 13 && callbackData[:13] == "back_to_cat:":
		categoryID := callbackData[13:]
		handleCategorySelect(b, callback.Message.Chat.ID, callback.From.ID, categoryID)

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
			msg := tgbotapi.NewMessage(chatID, "Категория: "+category.Name+"\n\nВыберите подкатегорию:")

			buttons := make([][]tgbotapi.InlineKeyboardButton, 0)

			for _, sub := range subcategories {
				buttons = append(buttons, []tgbotapi.InlineKeyboardButton{
					tgbotapi.NewInlineKeyboardButtonData(sub.Name, "subcategory:"+strconv.Itoa(sub.ID)),
				})
			}

			buttons = append(buttons, []tgbotapi.InlineKeyboardButton{
				tgbotapi.NewInlineKeyboardButtonData("➕ Добавить карточку", "add_card:"+categoryID),
				tgbotapi.NewInlineKeyboardButtonData("📋 Показать все", "show_cards:"+categoryID),
			})

			buttons = append(buttons, []tgbotapi.InlineKeyboardButton{
				tgbotapi.NewInlineKeyboardButtonData("🔧 Добавить подкатегорию", "add_subcategory:"+categoryID),
			})

			buttons = append(buttons, []tgbotapi.InlineKeyboardButton{
				tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", "back_main"),
			})

			msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(buttons...)
			b.bot.Send(msg)
			return
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

	card := &database.Card{
		PhotoFileID:   photo.FileID,
		CategoryID:    state.CategoryID,
		SubcategoryID: nil,
		UserID:        userID,
	}

	if state.SubcategoryID > 0 {
		card.SubcategoryID = &state.SubcategoryID
		var categoryID int
		err := b.DB.QueryRow("SELECT category_id FROM subcategories WHERE id = ?", state.SubcategoryID).Scan(&categoryID)
		if err == nil {
			card.CategoryID = categoryID
		}
	}

	err := storage.AddCard(b.DB, card)
	if err != nil {
		log.Println("Ошибка сохранения карточки:", err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка при сохранении карточки")
		b.bot.Send(msg)
		ClearUserState(userID)
		return
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "✅ Карточка успешно добавлена!")
	b.bot.Send(msg)

	ClearUserState(userID)

	msg2 := tgbotapi.NewMessage(message.Chat.ID, "👋 Выберите категорию:")
	msg2.ReplyMarkup = createMainKeyboard(b.DB, userID)
	b.bot.Send(msg2)
}

func handleText(b *Bot, message *tgbotapi.Message) {
	userID := message.From.ID
	text := message.Text

	state, exists := GetUserState(userID)
	if !exists {
		return
	}

	switch state.State {
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

func handleReset(b *Bot, chatID int64, userID int64) {
	if err := database.ResetUserCategories(b.DB, userID); err != nil {
		log.Printf("Ошибка сброса категорий: %v", err)
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка сброса категорий")
		b.bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(chatID, "✅ Категории успешно сброшены!")
	b.bot.Send(msg)

	// Показываем главное меню
	handleStart(b, chatID, userID)
}

func handleSubcategorySelect(b *Bot, chatID int64, userID int64, subcategoryID string) {
	subcategoryIDInt, err := strconv.Atoi(subcategoryID)
	if err != nil {
		log.Printf("Ошибка преобразования subcategoryID: %s, ошибка: %v", subcategoryID, err)
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка идентификатора подкатегории")
		b.bot.Send(msg)
		return
	}

	var subcategory database.Subcategory
	err = b.DB.QueryRow("SELECT id, name, category_id, user_id FROM subcategories WHERE id = ? AND user_id = ?",
		subcategoryIDInt, userID).Scan(&subcategory.ID, &subcategory.Name, &subcategory.CategoryID, &subcategory.UserID)
	if err != nil {
		log.Println("DB error:", err)
		msg := tgbotapi.NewMessage(chatID, "❌ Подкатегория не найдена")
		b.bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(chatID, "Подкатегория: "+subcategory.Name)

	buttons := [][]tgbotapi.InlineKeyboardButton{
		{
			tgbotapi.NewInlineKeyboardButtonData("➕ Добавить карточку", "add_sub_card:"+subcategoryID),
			tgbotapi.NewInlineKeyboardButtonData("📋 Показать все", "show_sub_cards:"+subcategoryID),
		},
		{
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", "back_to_cat:"+strconv.Itoa(subcategory.CategoryID)),
		},
	}

	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(buttons...)
	b.bot.Send(msg)
}

func handleAddSubcategoryCard(b *Bot, chatID int64, userID int64, subcategoryID string) {
	subcategoryIDInt, err := strconv.Atoi(subcategoryID)
	if err != nil {
		log.Printf("Ошибка преобразования subcategoryID: %s, ошибка: %v", subcategoryID, err)
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка идентификатора подкатегории")
		b.bot.Send(msg)
		return
	}

	SetUserStateWithSubcategory(userID, "waiting_photo", 0, subcategoryIDInt)
	msg := tgbotapi.NewMessage(chatID, "📸 Пришлите фото для карточки:")
	b.bot.Send(msg)
}

func handleShowSubcategoryCards(b *Bot, chatID int64, userID int64, subcategoryID string) {
	subcategoryIDInt, err := strconv.Atoi(subcategoryID)
	if err != nil {
		log.Printf("Ошибка преобразования subcategoryID: %s, ошибка: %v", subcategoryID, err)
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка идентификатора подкатегории")
		b.bot.Send(msg)
		return
	}

	cards, err := storage.GetSubcategoryCards(b.DB, userID, subcategoryIDInt)
	if err != nil {
		log.Println("Ошибка получения карточек:", err)
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка при получении карточек")
		b.bot.Send(msg)
		return
	}

	if len(cards) == 0 {
		msg := tgbotapi.NewMessage(chatID, "📭 В этой подкатегории пока нет карточек")
		b.bot.Send(msg)
	} else {
		for _, card := range cards {
			photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileID(card.PhotoFileID))
			b.bot.Send(photo)
		}
	}

	var subcategory database.Subcategory
	err = b.DB.QueryRow("SELECT id, name, category_id, user_id FROM subcategories WHERE id = ? AND user_id = ?",
		subcategoryIDInt, userID).Scan(&subcategory.ID, &subcategory.Name, &subcategory.CategoryID, &subcategory.UserID)
	if err == nil {
		msg := tgbotapi.NewMessage(chatID, "Вы можете:")
		buttons := [][]tgbotapi.InlineKeyboardButton{
			{
				tgbotapi.NewInlineKeyboardButtonData("➕ Добавить карточку", "add_sub_card:"+subcategoryID),
				tgbotapi.NewInlineKeyboardButtonData("📋 Показать все", "show_sub_cards:"+subcategoryID),
			},
			{
				tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", "back_to_cat:"+strconv.Itoa(subcategory.CategoryID)),
			},
		}
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(buttons...)
		b.bot.Send(msg)
	}
}

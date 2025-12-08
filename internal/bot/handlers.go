package bot

import (
	bookinfo "HIGH_PR/bookInfo"
	"HIGH_PR/gl"
	booktags "HIGH_PR/internal/repository/postgres/bookTags"
	"context"
	"fmt"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"

	"github.com/gotd/td/telegram/message/markup"
	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/telegram/uploader" // для uploader.NewUploader
	"strings"
	"time"

	//"os"
	"path/filepath"
	"strconv"
	//"time"
)

// Обработчик который делает запрос к бд и отдает все книги в формате (Название, Авторы, Описание, тэги)
func (b *Bot) handleShow(ctx context.Context, e tg.Entities, msg *tg.Message) {
	_, user, peer, err := getInfo(e, msg)
	b.logger.Printf("📨 /show от %s %s (@%s, ID:%d)",
		user.FirstName, user.LastName, user.Username, user.ID)

	if err != nil {
		b.logger.Println(err)
		return
	}

	// ВЫЗЫВАЕМ СЕРВИС (вместо прямого обращения к БД)
	books, err := b.bookService.GetAllBooks(ctx)
	if err != nil {
		b.logger.Println("Ошибка получения книг:", err)
		return
	}
	sender := message.NewSender(b.client.API()).To(peer)
	if len(books) != 0 {

		// Передаём готовые книги в функцию форматирования
		err = b.ShowBooksMessage(ctx, sender, books)
		if err != nil {
			b.logger.Println(err)
		}

	} else {
		sender.Text(ctx, "Книги не найдены!")
	}
}

// ATTENTION
func (b *Bot) handleShowWithID(ctx context.Context, e tg.Entities, msg *tg.Message) {
	_, user, peer, err := getInfo(e, msg)
	messageText := strings.TrimSpace(msg.Message)
	// Получение запроса в формате /show_1
	idStr := messageText[6:] // начинаем с 6, чтобы пропустить "/show_"
	id, err := strconv.Atoi(idStr)
	if err != nil {
		b.logger.Printf("❌ Ошибка парсинга ID: %v", err)
		// Отправить пользователю сообщение об ошибке
		return
	}

	b.logger.Printf("📨 /show_%d от %s %s (@%s, ID:%d)",
		id, user.FirstName, user.LastName, user.Username, user.ID)

	book, err := b.bookService.BookWithID(ctx, id)
	sender := message.NewSender(b.client.API()).To(peer)
	if err != nil {
		b.logger.Println("Ошибка при выполнения запроса к бд. ERR = ", err)
		sender.Text(ctx, fmt.Sprintf("Ошибка выполнения запроса\nERR=%s", err))
		return
	}
	err = ShowBookWithIDMessage(ctx, b.client, peer, book)
	if err != nil {
		b.logger.Println("Ошибка в ShowBookWithIDMessage. ERR = ", err)
		sender.Text(ctx, fmt.Sprintf("ERR = %s", err))
	}

	b.logger.Println("Успешно отправлена инфа о книге ID =", id)

}

func (b *Bot) handleShowWithName(ctx context.Context, e tg.Entities, msg *tg.Message) {

	_, user, peer, err := getInfo(e, msg)

	b.logger.Printf("📨 /showWithName от %s %s (@%s, ID:%d)",
		user.FirstName, user.LastName, user.Username, user.ID)
	if err != nil {
		b.logger.Println(err)
		return
	}
	text := msg.Message
	nameBook := strings.TrimPrefix(text, "/WithName")
	nameBook = strings.TrimSpace(nameBook)
	sender := message.NewSender(b.client.API()).To(peer)
	if len(nameBook) != 0 {
		res, err := bookinfo.SearchBooks(nameBook)
		if err != nil {
			b.logger.Println("Ошибка запроса к Google API Books: ", err)

		}
		if res.Title == "" {
			b.logger.Println("Не найдено")
			sender.Text(ctx, "Книга не найдена 🫠🫠🫠\nПроверте название 🍄🍄🍄")
			return

		}

		b.logger.Println("!!! Запрос на получение книги: ", nameBook)

		sender.Text(ctx, fmt.Sprintf("Название: %s\n\nАвторы: %v\n\nОписание: %s\n", res.Title, res.Authors, res.Description))
	} else {
		b.logger.Println("Не указано имя!")
		sender.Text(ctx, "Укажите имя!")
	}

}

// ATTENTION
// Функция обрабатывающая команду которая отдает книгу для скичивания (в нее передаеться ID, (также сделаю кликабельный названия книг в команде show() который будут делать запрос со скачиванием и передавать ID))
func (b *Bot) handleDownloadBook(ctx context.Context, e tg.Entities, msg *tg.Message) {
	// 1. Получаем информацию об отправителе
	_, user, peer, err := getInfo(e, msg)
	if err != nil {
		b.logger.Println("Ошибка получения инфо:", err)
		return
	}

	// 2. Парсим ID из сообщения
	messageText := strings.TrimSpace(msg.Message)
	// Ожидаем формат "/download_123"
	if len(messageText) < 10 {
		return
	}
	idStr := messageText[10:] // Отрезаем "/download_"
	id, err := strconv.Atoi(idStr)
	if err != nil {
		b.logger.Printf("❌ Ошибка парсинга ID: %v", err)
		// Тут можно отправить sender.Text(ctx, "Неверный ID книги")
		return
	}

	b.logger.Printf("📨 Запрос книги ID:%d от %s (ID:%d)", id, user.FirstName, user.ID)
	err = b.bookService.AddDownloadCountWithID(ctx, id)
	if err != nil {
		b.logger.Println("Ошибка инкремента скачиваний: ", err)
	}

	// Инициализируем сендер
	sender := message.NewSender(b.client.API()).To(peer)

	// 3. Получаем путь к файлу (твоя логика)
	filePath, err := b.bookService.GetFileBookWithID(ctx, id)
	if err != nil {
		b.logger.Println("Ошибка получения файла:", err)
		sender.Text(ctx, "Извините, файл не найден.")
		return
	}

	// 4. Загружаем файл в Telegram
	// uploader.NewUploader разбивает файл на части и отправляет их
	u := uploader.NewUploader(b.client.API())

	b.logger.Println("Начинаю загрузку файла:", filePath)
	inputFile, err := u.FromPath(ctx, filePath)
	if err != nil {
		b.logger.Println("Ошибка загрузки (upload):", err)
		sender.Text(ctx, "Ошибка при загрузке файла.")
		return
	}

	// 5. Подготавливаем InputMediaUploadedDocument [web:7][web:10]
	// Это ключевой момент: конвертируем загруженный файл в медиа-объект
	// Обязательно нужно указать имя файла через Attributes, иначе он придет как "file" без расширения
	fileName := filepath.Base(filePath)

	media := &tg.InputMediaUploadedDocument{
		File:     inputFile,
		MimeType: "application/pdf", // Желательно определять реально (например, "application/pdf")
		Attributes: []tg.DocumentAttributeClass{
			&tg.DocumentAttributeFilename{FileName: fileName}, // Чтобы у файла было имя
		},
		ForceFile: true, // Форсируем отправку именно как файл (документ)
	}

	// 6. Отправляем через метод Media(), а не Document()
	// Document() ждет tg.InputDocument (существующий файл), а мы шлем tg.InputMedia (новый контент)
	if _, err := sender.Media(ctx, message.Media(media)); err != nil {
		b.logger.Println("Ошибка отправки медиа:", err)
		sender.Text(ctx, fmt.Sprintf("Не удалось отправить файл: %s", err))
		return
	}

	b.logger.Println("Файл успешно отправлен!")
}

// handleStart обрабатывает команду /start
func (b *Bot) handleStart(ctx context.Context, e tg.Entities, msg *tg.Message) {
	_, user, peer, err := getInfo(e, msg)

	if err != nil {
		b.logger.Println(err)
	}

	b.logger.Printf("📨 /start от %s %s (@%s, ID:%d)",
		user.FirstName,
		user.LastName,
		user.Username,
		user.ID)

	// 6. Отправляем ответ

	//WARNING

	//Задокументированный код снизу представляет работу с сырыми сообщениями (в данном примере предсавленна отправка текста и и отправка специального объекта для скрытия клавиатуры (не inline!))

	// // Генерируем random ID (можно использовать time.Now().UnixNano())
	// randomID := time.Now().UnixNano()
	//
	// // Отправляем сообщение с удалением клавиатуры
	// _, err = b.client.API().MessagesSendMessage(ctx, &tg.MessagesSendMessageRequest{
	// 	Peer:     peer,
	// 	Message:  fmt.Sprintf("Привет, %s! 👋", user.FirstName),
	// 					    RandomID: randomID,  // ВОТ ЭТО ОБЯЗАТЕЛЬНО!
	// 					    ReplyMarkup: &tg.ReplyKeyboardHide{
	// 						    Selective: false,
	// 					    },
	// })

	_, err = b.client.API().MessagesSendMessage(ctx, &tg.MessagesSendMessageRequest{
		Peer:     peer, // получатель
		Message:  fmt.Sprintf("Привет, %s! 👋", user.FirstName),
		RandomID: time.Now().UnixNano(), // всегда уникальный
	})
	u := uploader.NewUploader(b.client.API()).
	WithPartSize(512 * 1024). // 512 KB (стандартный чанк)
	WithThreads(4)

	// 2. Указываем путь к файлу
	// Важно: файл должен быть в формате .webp (для обычных стикеров)
	filePath := "/home/magamed/Рабочий стол/MyPet/TG/HIGH_PR/sticker/hello.webp"

	// 3. Загружаем файл на сервера Telegram
	// FromPath сам откроет файл и загрузит его
	upload, err := u.FromPath(ctx, filePath)
	if err != nil {
		b.logger.Println("Ошибка загрузки файла:", err)

	}
	sender := message.NewSender(b.client.API()).To(peer)
	// 4. Отправляем загруженный файл именно как СТИКЕР
	// Метод UploadedSticker берет уже загруженный файл и делает из него сообщение
	_, err = sender.UploadedSticker(ctx, upload)
	if err != nil {
		b.logger.Println("Ошибка отправки стикера:", err)

	}

	if err != nil {
		b.logger.Println("Ошибка отправки сообщения с эффектом:", err)
	}

}

// handleHelp обрабатывает команду /help
func (b *Bot) handleHelp(ctx context.Context, e tg.Entities, msg *tg.Message) {
	_, user, peer, err := getInfo(e, msg)

	if err != nil {
		b.logger.Println(err)
	}

	b.logger.Printf("📨 /help от %s %s (@%s, ID:%d)",
			user.FirstName,
		 user.LastName,
		 user.Username,
		 user.ID)
	sender := message.NewSender(b.client.API()).To(peer)
	text := b.SendHelpMessage(ctx)
	_, err = sender.StyledText(ctx,text...)
	if err != nil {
		b.logger.Println("Ошибка при отправке сообщения! ",err)
	}

}

// handleAddBook обрабатывает команду /add
func (b *Bot) handleAddBook(ctx context.Context, e tg.Entities, msg *tg.Message, update *tg.UpdateNewMessage) {
	txt := strings.TrimSpace(msg.Message)
	_, user, peer, err := getInfo(e, msg)
	b.logger.Printf("📨 /add от %s %s (@%s, ID:%d)",
		user.FirstName, user.LastName, user.Username, user.ID)
	if err != nil {
		b.logger.Println(err)
		return
	}
	sender := message.NewSender(b.client.API()).To(peer)
	media, ok := msg.Media.(*tg.MessageMediaDocument)
	if !ok {
		b.logger.Println("Команда /add без файла!")
		_, err = sender.Text(ctx, "Введите комманду /add и прикрепите файл!")
		if err != nil {
			b.logger.Printf("Ошибка отправки: %v", err)
		}
		return
	}
	doc, ok := media.Document.(*tg.Document)
	if !ok {
		b.logger.Println("не удалось получить документ")
	}

	if txt == "/add" {
		fullName := GetDocumentName(doc)
		name := DeleteType(fullName)
		info, err := bookinfo.SearchBooks(name)
		fileType := ExtractFileFormat(fullName)
		if err != nil {
			b.logger.Println("Ошибка загрузки информации о книге!")
			sender.Text(ctx, "Ошибка загрузки информации о книги из Google Book API")
			sender.Text(ctx, "Попробуйте ввести название книги в ручную! (/add Название книги...+файл)")
			return
		}
		langTag, otherTag, err := bookinfo.ParseMetadataFromInfo(info.Title,info.Description)
		if err != nil {
			b.logger.Println("Ошибка парса тэга")
		}

		err = b.bookService.AddBook(ctx, booktags.BookWithTags{
			B: booktags.Book{
				Title:       info.Title,
				Authors:     info.Authors,
				Description: info.Description,
				TextSnippet: info.TextSnippet,
				FileSize:    doc.Size,
				Img:         info.Img,
				FileType:    fileType,
				FilePath:    gl.DefaultSaveBook + "/" + fullName,
				AddedBy:     user.Username,
				AddedAt:     time.Now().Truncate(time.Second),
			},
			T: booktags.Tag{
				Lang:            info.Lang,
				ProgrammingLang: []string{langTag},
				OtherTag:        []string{otherTag},
			},
		})
		if err != nil {
			b.logger.Println("Ошибка добавления книги в бд:", err)
			sender.Text(ctx, fmt.Sprintf("ERR=%s", err))
			return
		}

		err = b.DownloadFile(ctx, media)
		if err != nil {
			b.logger.Println("Ошибка загрузки файла: ", err)
			_, err = sender.Text(ctx, fmt.Sprintf("Ошибка при загрузке файла!\nError: %s", err))
			if err != nil {
				b.logger.Printf("Ошибка отправки: %v", err)
			}
			return
		}
		_, err = sender.Text(ctx,"Файл успешно сохранен!")
		if err != nil {
			b.logger.Printf("Ошибка отправки: %v", err)
		}

	} else {
		nameBook := strings.TrimPrefix(txt, "/add")

		fullName := GetDocumentName(doc)
		info, err := bookinfo.SearchBooks(nameBook)
		fileType := ExtractFileFormat(fullName)

		if err != nil {
			b.logger.Println("Ошибка загрузки информации о книге!")
			sender.Text(ctx, "Ошибка загрузки информации о книги из Google Book API")
		}

		langTag, otherTag, err := bookinfo.ParseMetadataFromInfo(info.Title,info.Description)
		if err != nil {
			b.logger.Println("Ошибка парса тэга")
		}

		err = b.bookService.AddBook(ctx, booktags.BookWithTags{
			B: booktags.Book{
				Title:       info.Title,
				Authors:     info.Authors,
				Description: info.Description,
				TextSnippet: info.TextSnippet,
				FileSize:    doc.Size,
				Img:         info.Img,
				FileType:    fileType,
				FilePath:    gl.DefaultSaveBook + "/" + fullName,
				AddedBy:     user.Username,
				AddedAt:     time.Now().Truncate(time.Second),
			},
			T: booktags.Tag{
				Lang:            info.Lang,
				ProgrammingLang: []string{langTag},
				OtherTag:        []string{otherTag},
			},
		})

		if err != nil {
			b.logger.Println("Ошибка добавления книги в бд:", err)
			sender.Text(ctx, fmt.Sprintf("ERR=%s", err))
			return
		}

		err = b.DownloadFile(ctx, media)
		if err != nil {
			b.logger.Println("Ошибка загрузки файла: ", err)
			_, err = sender.Text(ctx, fmt.Sprintf("Ошибка при загрузке файла!\nError: %s", err))
			if err != nil {
				b.logger.Printf("Ошибка отправки: %v", err)
			}
			return
		}
		_, err = sender.Text(ctx, fmt.Sprint("Файл успешно сохранен!"))
		if err != nil {
			b.logger.Printf("Ошибка отправки: %v", err)
		}

	}

}

// handleSearch обрабатывает поиск
func (b *Bot) handleSearch(ctx context.Context, e tg.Entities, msg *tg.Message) {
	// ... логика для поиска ...
}
func (b *Bot) handleAdmin(ctx context.Context, e tg.Entities, msg *tg.Message) {

	_, user, peer, err := getInfo(e, msg)

	if err != nil {
		b.logger.Println(err)
	}

	b.logger.Printf("📨 /admin от %s %s (@%s, ID:%d)",
		user.FirstName,
		user.LastName,
		user.Username,
		user.ID)

	sender := message.NewSender(b.client.API()).To(peer)
	adminID, _ := strconv.ParseInt(gl.AdminID, 10, 64)

	if user.ID != adminID {
		_, err = sender.Text(ctx, fmt.Sprintf("Привет, %s! 👋\nДоступ Запрещен!", user.Username))
		if err != nil {
			b.logger.Printf("Ошибка отправки: %v", err)
		}
	} else {
		_, err := sender.
			Markup(markup.InlineRow(
				markup.Callback("All_Log", []byte("FileLog")),
				markup.Callback("Last_Log", []byte("LastLog")),
			)).
			Text(ctx, fmt.Sprintf("Привет мой повелитель %s", user.Username))

		if err != nil {
			b.logger.Printf("Ошибка отправки: %v", err)
		}
		//Вызываем обработчик универсал

	}
}

// WARNING
// Функция для получение id, юзера и peer
func getInfo(e tg.Entities, msg *tg.Message) (int64, *tg.User, tg.InputPeerClass, error) {
	var userID int64
	var user *tg.User
	var ok bool

	entities := peer.NewEntities(e.Users, e.Chats, e.Channels)

	// 1. Сначала пробуем FromID (входящие сообщения)
	if fromID, hasFromID := msg.GetFromID(); hasFromID {
		if peerUser, isPeerUser := fromID.(*tg.PeerUser); isPeerUser {
			userID = peerUser.UserID
		}
	}

	// 2. Если FromID пустой, используем PeerID (для личных чатов)
	if userID == 0 {
		if peerID, ok := msg.PeerID.(*tg.PeerUser); ok {
			userID = peerID.UserID
		}
	}

	// 3. Если всё ещё не нашли, это групповой чат или ошибка
	if userID == 0 {

		return 0, nil, nil, fmt.Errorf("не удалось определить пользователя")
	}

	// 4. Получаем данные пользователя
	user, ok = entities.User(userID)
	if !ok {
		return 0, nil, nil, fmt.Errorf("пользователь %d не найден в Entities", userID)
	}

	peer, err := entities.ExtractPeer(msg.PeerID)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("ошибка ExtractPeer: %v", err)
	}
	return userID, user, peer, nil
}

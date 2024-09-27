package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"simple-telegram-bot/model"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

const (
	epsURL   = "eps"
	NotFound = "N/A"
	Start    = "/start"
	Location = "📍 Location"

	EN = "🇬🇧 En"
	UK = "🇬🇧 Uk"
	DE = "🇩🇪 De"
	FR = "🇫🇷 Fr"
	PL = "🇵🇱 Pl" // Polish

	DK = "🇩🇰 Dk" // Denmark
	DA = "🇩🇰 Da" // Danish
	ES = "🇪🇸 Es" // Spanish
	PT = "🇵🇹 Pt" // Portuguese
	RU = "🇷🇺 Ru" // Russian

	FI = "🇫🇮 Fi" // Finnish
	NO = "🇳🇴 No" // Norwegian
	IS = "🇮🇸 Is" // Icelandic
	SV = "🇸🇪 Sv" // Swedish
	SE = "🇸🇪 Se" // Sweden

	IT = "🇮🇹 It" // Italian
	BS = "🇧🇦 Bs" // Bosnian
	SR = "🇷🇸 Sr" // Serbian
	RO = "🇷🇴 Ro" // Romanian

	AL = "🇦🇱 Al" // Albania
	AD = "🇦🇩 Ad" // Andorra
	AT = "🇦🇹 At" // Austria
	BE = "🇧🇪 Be" // Belgium

	BA = "🇧🇦 Ba" // Bosnia and Herzegovina
	BG = "🇧🇬 Bg" // Bulgaria
	HR = "🇭🇷 Hr" // Croatia
	CY = "🇨🇾 Cy" // Cyprus

	CZ = "🇨🇿 Cz" // Czechia
	EE = "🇪🇪 Ee" // Estonia
	GR = "🇬🇷 Gr" // Greece
	HU = "🇭🇺 Hu" // Hungary

	IE = "🇮🇪 Ie" // Ireland
	LV = "🇱🇻 Lv" // Latvia
	LI = "🇱🇮 Li" // Liechtenstein
	LT = "🇱🇹 Lt" // Lithuania

	LU = "🇱🇺 Lu" // Luxembourg
	MT = "🇲🇹 Mt" // Malta
	MD = "🇲🇩 Md" // Moldova
	MC = "🇲🇨 Mc" // Monaco

	ME = "🇲🇪 Me" // Montenegro
	MK = "🇲🇰 Mk" // North Macedonia
	SM = "🇸🇲 Sm" // San Marino
	SK = "🇸🇰 Sk" // Slovakia

	UA = "🇺🇦 Ua" // Ukraine
	VA = "🇻🇦 Va" // Vatican City
	SI = "🇸🇮 Si" // Slovenia
	TR = "🇹🇷 Tr" // Turkiye

	Language       = "🌐 Language"
	GetLastResult  = "📈 Get Last Results"
	SetSearchQuery = "🔎 Set Your Search Query"
	StartCrawler   = "🪲 Start Crawler"
	GetAllSQs      = "Get Search Queries"
	ExtractCSV     = "📄 Export CSV File"
	// CancelAJob     = "Cancel a Job"
	NextPage      = "Next Page"
	Back          = "Back"
	APIToken      = "7470825262:AAHW4wBh7LEk1oCEPsBMkdyQeCeadwfl2Ew" //"7053131148:AAG3NtM0ZJFxHEiGRQoUXeKlhDLqfVj5x78" //"6704135678:AAH2SGETz7tSKY5NsQ-vv2-zO7tj-XZaKAk"
	TGBotPassword = "0a164SO55RD0c0nb:@}s"
	UnknownInput  = "unknown_input"
	Empty         = ""
)

var (
	authUsersMap = map[string]bool{
		"gopher_dev1997":   false,
		"m_heydari4883":    false,
		"CrmVercell":       false,
		"Gharazi69":        false,
		"EsmaeilAlinezhad": false,
	}
	lastInput               string
	SQID                    = 0
	getLastResultPage       = 0
	getAllSearchQueriesPage = 0
	sq                      = model.SearchQueryRequest{}
)

var (
	emptyKey = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(`¯\_(ツ)_/¯`)),
	)
	lastResultKey = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(NextPage)),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(Back)),
	)
	helpKeys = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(SetSearchQuery)),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(StartCrawler)),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(GetLastResult)),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(ExtractCSV)),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(GetAllSQs),
			// tgbotapi.NewKeyboardButton(CancelAJob),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(Language),
			tgbotapi.NewKeyboardButton(Location),
		),
	)

	langKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(EN, "English"),
			tgbotapi.NewInlineKeyboardButtonData(DE, "German"),
			tgbotapi.NewInlineKeyboardButtonData(FR, "French"),
			tgbotapi.NewInlineKeyboardButtonData(PL, "Polish"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(DA, "Danish"),
			tgbotapi.NewInlineKeyboardButtonData(ES, "Spanish"),
			tgbotapi.NewInlineKeyboardButtonData(PT, "Portuguese"),
			tgbotapi.NewInlineKeyboardButtonData(RU, "Russian"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(FI, "Finnish"),
			tgbotapi.NewInlineKeyboardButtonData(NO, "Norwegian"),
			tgbotapi.NewInlineKeyboardButtonData(IS, "Icelandic"),
			tgbotapi.NewInlineKeyboardButtonData(SV, "Swedish"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(IT, "Italian"),
			tgbotapi.NewInlineKeyboardButtonData(BS, "Bosnian"),
			tgbotapi.NewInlineKeyboardButtonData(SR, "Serbian"),
			tgbotapi.NewInlineKeyboardButtonData(RO, "Romanian"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(AL, "Albanian"),
			tgbotapi.NewInlineKeyboardButtonData(AD, "Catalan"),
			tgbotapi.NewInlineKeyboardButtonData(HR, "Croatian"),
			tgbotapi.NewInlineKeyboardButtonData(BG, "Bulgarian"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(TR, "Turkish"),
			tgbotapi.NewInlineKeyboardButtonData(GR, "Greek"),
			tgbotapi.NewInlineKeyboardButtonData(CZ, "Czech"),
			tgbotapi.NewInlineKeyboardButtonData(EE, "Estonian"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(HU, "Hungarian"),
			tgbotapi.NewInlineKeyboardButtonData(LV, "Latvian"),
			tgbotapi.NewInlineKeyboardButtonData(IE, "Irish"),
			tgbotapi.NewInlineKeyboardButtonData(LT, "Lithuanian"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(MT, "Maltese"),
			tgbotapi.NewInlineKeyboardButtonData(MK, "Macedonian"),
			tgbotapi.NewInlineKeyboardButtonData(SK, "Slovak"),
			tgbotapi.NewInlineKeyboardButtonData(SI, "Slovenian"),
		),
		// New LANGUAGEs
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(UA, "Ukrainian"),
		// 	tgbotapi.NewInlineKeyboardButtonData(MK, "Macedonian"),
		// 	tgbotapi.NewInlineKeyboardButtonData(SK, "Slovak"),
		// 	tgbotapi.NewInlineKeyboardButtonData(SI, "Slovenian"),
		),
	)

	locKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(UK, "England"),
			tgbotapi.NewInlineKeyboardButtonData(DE, "Germany"),
			tgbotapi.NewInlineKeyboardButtonData(FR, "France"),
			tgbotapi.NewInlineKeyboardButtonData(PL, "Poland"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(DK, "Denmark"),
			tgbotapi.NewInlineKeyboardButtonData(ES, "Spain"),
			tgbotapi.NewInlineKeyboardButtonData(PT, "Portugal"),
			tgbotapi.NewInlineKeyboardButtonData(TR, "Turkiye"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(FI, "Finland"),
			tgbotapi.NewInlineKeyboardButtonData(NO, "Norway"),
			tgbotapi.NewInlineKeyboardButtonData(IS, "Iceland"),
			tgbotapi.NewInlineKeyboardButtonData(SE, "Sweden"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(IT, "Italy"),
			tgbotapi.NewInlineKeyboardButtonData(BS, "Bosnia"),
			tgbotapi.NewInlineKeyboardButtonData(SR, "Serbia"),
			tgbotapi.NewInlineKeyboardButtonData(RO, "Romania"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(AL, "Albania"),
			tgbotapi.NewInlineKeyboardButtonData(AD, "Andorra"),
			tgbotapi.NewInlineKeyboardButtonData(AT, "Austria"),
			tgbotapi.NewInlineKeyboardButtonData(BE, "Belgium"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(BA, "Bosnia and Herzegovina"),
			tgbotapi.NewInlineKeyboardButtonData(BG, "Bulgaria"),
			tgbotapi.NewInlineKeyboardButtonData(HR, "Croatia"),
			tgbotapi.NewInlineKeyboardButtonData(CY, "Cyprus"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(CZ, "Czechia"),
			tgbotapi.NewInlineKeyboardButtonData(EE, "Estonia"),
			tgbotapi.NewInlineKeyboardButtonData(GR, "Greece"),
			tgbotapi.NewInlineKeyboardButtonData(HU, "Hungary"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(IE, "Ireland"),
			tgbotapi.NewInlineKeyboardButtonData(LV, "Latvia"),
			tgbotapi.NewInlineKeyboardButtonData(LI, "Liechtenstein"),
			tgbotapi.NewInlineKeyboardButtonData(LT, "Lithuania"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(LU, "Luxembourg"),
			tgbotapi.NewInlineKeyboardButtonData(MT, "Malta"),
			tgbotapi.NewInlineKeyboardButtonData(MD, "Moldova"),
			tgbotapi.NewInlineKeyboardButtonData(MC, "Monaco"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(ME, "Montenegro"),
			tgbotapi.NewInlineKeyboardButtonData(MK, "North Macedonia"),
			tgbotapi.NewInlineKeyboardButtonData(SM, "San Marino"),
			tgbotapi.NewInlineKeyboardButtonData(SK, "Slovakia"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(UA, "Ukraine"),
			tgbotapi.NewInlineKeyboardButtonData(VA, "Vatican City"),
			tgbotapi.NewInlineKeyboardButtonData(SI, "Slovenia"),
		),
	)
)

func main() {
	bot, err := tgbotapi.NewBotAPI(APIToken)
	if err != nil {
		log.Panic(err)
	}

	//bot.Debug = true
	client := &http.Client{}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	// Loop through each update.
	for update := range updates {
		// Check if we've gotten a message update.
		if update.Message != nil {
			messageText := update.Message.Text
			messages := make([]tgbotapi.Chattable, 0)
			userName := update.Message.Chat.UserName
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, messageText)
			if !isUserAuthenticated(userName) {
				logrus.Infof("an unauthorized loggin attemption by %q user id", userName)
				continue
			}
			if !isUserAuthorized(userName) {
				switch {
				case lastInput == Start:
					if update.Message.Text == TGBotPassword {
						authorizeUser(userName)
						msg.ReplyMarkup = helpKeys
						msg.Text = "🔓 You are authenticated"
					} else {
						msg.Text = "🔒 Password is incorrect. Try again"
						// msg.Text = "🔑 Welcome to EPS Crawler Bot.\nTo use this bot, please provide your password:"
					}
				case messageText == Start:
					lastInput = Start
					msg.Text = "🔑 Welcome to EPS Crawler Bot.\nTo use this bot, please provide your password:"
				default:
					msg.Text = "⚠️ Maybe server just restarted.\nPlease clear search history and /start the bot again."
					msg.ReplyMarkup = emptyKey
				}
				if _, err = bot.Send(msg); err != nil {
					logrus.Info("failed to send message: %w", err)
				}
				continue
			} else {
				msg.ReplyMarkup = helpKeys
			}

			switch lastInput {
			case UnknownInput:
			case Start:
				messages = append(messages, msg)
			case SetSearchQuery:
				sq.Query = messageText
				msg.Text = "Your search query set."
				messages = append(messages, msg)
				lastInput = Empty
			case GetLastResult:
				sqid, err := strconv.Atoi(messageText)
				if err != nil {
					msg.Text = "Search query ID must be a number"
					break
				}
				SQID = sqid
				if SQID == 0 {
					msg.Text = "Search query ID cannot be zero"
					break
				}
				msgs, err := GetLastResults(client, update.Message.Chat.ID, SQID, getLastResultPage)
				if err != nil {
					switch {
					case errors.Is(err, sql.ErrNoRows):
						msg.Text = "No record found.\nUse 'Back' button to back to the Home page."
					default:
						logrus.Info(err)
						msg.Text = "Failed to perform your action, please try again later."
					}
					msg.ReplyMarkup = lastResultKey
					messages = append(messages, msg)
					break
				}
				msg.Text = fmt.Sprintf("Your search results for SQID: %d - page: %d", SQID, getLastResultPage+1)
				msg.ReplyMarkup = lastResultKey
				messages = append(messages, msg)
				for _, m := range msgs {
					messages = append(messages, m)
				}
				lastInput = GetLastResult
			case ExtractCSV:
				sqid, err := strconv.Atoi(messageText)
				if err != nil {
					msg.Text = "Search query ID must be a number"
					break
				}
				SQID = sqid
				if SQID == 0 {
					msg.Text = "Search query ID cannot be zero"
					break
				}
				docMsg, err := GetCSVFile(client, update.Message.Chat.ID, SQID)
				if err != nil {
					logrus.Info(err)
					msg.Text = "Failed to perform your action, please try again later."
					break
				}
				messages = append(messages, docMsg)
				msg.Text = Empty
				lastInput = Empty
				// case CancelAJob:
				// 	sqid, err := strconv.Atoi(messageText)
				// 	if err != nil {
				// 		msg.Text = "Search query ID must be a number"
				// 		break
				// 	}
				// 	SQID = sqid
				// 	if SQID == 0 {
				// 		msg.Text = "Search query ID cannot be zero"
				// 		break
				// 	}
				// 	res, err := CancelASearchQuery(client, SQID)
				// 	if err != nil {
				// 		logrus.Info(err)
				// 		if res == "" {
				// 			res = "Failed to perform your action, please try again later."
				// 		}
				// 	}
				// 	msg.Text = res
				// 	lastInput = Empty
				// 	messages = append(messages, msg)
			}

			// Construct a new message from the given chat ID and containing
			// the text that we received.
			// If the message was open, add a copy of our numeric keyboard.
			switch messageText {
			case Start:
				lastInput = Start
				if isUserAuthorized(userName) {
					msg.Text = "🔓 You are already authenticated.\nUse the provided keys to work with EPS Bot."
				} else {
					msg.Text = "🔑 Welcome to EPS Crawler Bot.\nTo use this bot, please provide your password:"
				}
				messages = append(messages, msg)
			case Location:
				lastInput = Location
				msg.Text = "📍 Select Your Location: "
				msg.ReplyMarkup = locKeyboard
				messages = append(messages, msg)
			case Language:
				lastInput = Language
				msg.Text = "🌐 Select Your Language: "
				msg.ReplyMarkup = langKeyboard
				messages = append(messages, msg)
			case SetSearchQuery:
				lastInput = SetSearchQuery
				msg.Text = "Please provide your search query to let the crawler starts it's work:"
				messages = append(messages, msg)
			case GetLastResult:
				lastInput = GetLastResult
				msg.Text = "Please provide your search query ID to get the latest results:"
				messages = append(messages, msg)
			case GetAllSQs:
				lastInput = GetAllSQs
				msgs, err := GetAllSearchQueries(client, update.Message.Chat.ID, getAllSearchQueriesPage)
				if err != nil {
					logrus.Info(err)
					msg.Text = "Failed to perform your action, please try again later."
					break
				}
				msg.Text = fmt.Sprintf("Your search queries - page: %d", getAllSearchQueriesPage+1)
				msg.ReplyMarkup = lastResultKey
				messages = append(messages, msg)
				for _, m := range msgs {
					messages = append(messages, m)
				}
				// lastInput = Empty
			case StartCrawler:
				lastInput = StartCrawler
				m, err := StartTheCrawler(client)
				if err != nil {
					logrus.Info(err)
					if m == "" {
						m = "Failed to perform your action, please try again later."
					}
				}
				lastInput = Empty
				msg.Text = m
				messages = append(messages, msg)
			case ExtractCSV:
				lastInput = ExtractCSV
				msg.Text = "Please provide your search query ID to send export it:"
				messages = append(messages, msg)
			// case CancelAJob:
			// 	lastInput = CancelAJob
			// 	msg.Text = "⚠️ NOTE: if you cancel a job you wont be able to reactivate it.\n\nPlease provide your search query ID to cancel it:"
			// 	messages = append(messages, msg)
			case NextPage:
				// lastInput = NextPage
				fmt.Println(lastInput)
				if lastInput == GetLastResult {
					getLastResultPage++
					msgs, err := GetLastResults(client, update.Message.Chat.ID, SQID, getLastResultPage)
					if err != nil {
						switch {
						case errors.Is(err, sql.ErrNoRows):
							msg.Text = "No record found.\nUse 'Back' button to back to the Home page."
						default:
							logrus.Info(err)
							msg.Text = "Failed to perform your action, please try again later."
						}
						msg.ReplyMarkup = lastResultKey
						messages = append(messages, msg)
						break
					}
					msg.Text = fmt.Sprintf("Your search results for SQID: %d - page: %d", SQID, getLastResultPage+1)
					msg.ReplyMarkup = lastResultKey
					messages = append(messages, msg)
					msg.ReplyMarkup = lastResultKey
					for _, m := range msgs {
						messages = append(messages, m)
					}
				}
				if lastInput == GetAllSQs {
					getAllSearchQueriesPage++
					msgs, err := GetAllSearchQueries(client, update.Message.Chat.ID, getAllSearchQueriesPage)
					if err != nil {
						switch {
						case errors.Is(err, sql.ErrNoRows):
							msg.Text = "No record found.\nUse 'Back' button to back to the Home page."
						default:
							logrus.Info(err)
							msg.Text = "Failed to perform your action, please try again later."
						}
						msg.ReplyMarkup = lastResultKey
						messages = append(messages, msg)
						break
					}
					msg.Text = fmt.Sprintf("Your search queries - page: %d", getAllSearchQueriesPage+1)
					msg.ReplyMarkup = lastResultKey
					messages = append(messages, msg)
					for _, m := range msgs {
						messages = append(messages, m)
					}
				}
			case Back:
				// lastInput = Back
				getLastResultPage = 0
				getAllSearchQueriesPage = 0
				msg.ReplyMarkup = helpKeys
				msg.Text = "Back to Home"
				messages = append(messages, msg)
				lastInput = Empty
			default:
				// lastInput = UnknownInput
			}
			for _, m := range messages {
				if _, err = bot.Send(m); err != nil {
					logrus.Info("failed to send message: %w", err)
				}
			}
		} else if update.CallbackQuery != nil {
			// Respond to the callback query, telling Telegram to show the user
			// a message with the data received.
			clbkQData := update.CallbackQuery.Data
			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, clbkQData)
			switch lastInput {
			case Language:
				callback.Text = fmt.Sprintf("🌐 %s language selected", clbkQData)
				sq.Language = languages[clbkQData]
				lastInput = Empty
			case Location:
				callback.Text = fmt.Sprintf("📍 %s location selected", clbkQData)
				sq.Location = locations[clbkQData]
				lastInput = Empty
			default:
				callback.Text = "I don't know what you mean, please use below buttons:"
			}
			if _, err := bot.Request(callback); err != nil {
				logrus.Info("bot.Request: %w", err)
			}
		}
	}
}

func GetLastResults(client *http.Client, msgChatID int64, sqID, page int) ([]tgbotapi.MessageConfig, error) {
	urlPath := fmt.Sprintf("http://%s:9999/api/v1/search/%d?page=%d", epsURL, sqID, page)
	req, err := http.NewRequest(http.MethodGet, urlPath, nil)
	if err != nil {
		return nil, fmt.Errorf("GetLastResults.NewRequest: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GetLastResults.Get: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GetLastResults.StatusCode: unintented status code: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	var serps []model.SERPResponse
	err = json.Unmarshal(bodyBytes, &serps)
	if err != nil {
		return nil, fmt.Errorf("GetLastResults.Unmarshal: %w", err)
	}
	tgMessages := make([]tgbotapi.MessageConfig, 0, len(serps))
	if len(serps) == 0 {
		msg := tgbotapi.NewMessage(msgChatID, "There is no more result in database.\nUse `Back` button to head to the Home.")
		tgMessages = append(tgMessages, msg)
		return tgMessages, sql.ErrNoRows
	}
	for _, serp := range serps {
		msg := tgbotapi.NewMessage(msgChatID, "")
		msg.ParseMode = tgbotapi.ModeHTML
		emails, keyWords := "", ""
		if len(serp.Emails) == 0 {
			emails = NotFound
		} else {
			emails = strings.Join(serp.Emails, "\n\t")
		}
		if len(serp.Keywords) == 0 {
			keyWords = NotFound
		} else {
			keyWords = strings.Join(serp.Keywords, "\n\t")
		}
		msg.Text = fmt.Sprintf(
			`<a href='%s'>%s</a>

%s

🏷️ Key Words:
%s

📧 Emails:
%s

☎️ Phones:
%s
`,
			serp.URL,
			serp.Title,
			serp.Description,
			keyWords,
			emails,
			preparePhones(serp.Phones),
		)
		tgMessages = append(tgMessages, msg)
	}
	if len(tgMessages) == 0 {
		return nil, sql.ErrNoRows
	}
	return tgMessages, nil
}
func preparePhones(ss []string) string {
	var result string
	if len(ss) == 0 {
		return NotFound
	}
	for _, s := range ss {
		result += fmt.Sprintf(`<a href="tel:%s">%s</a>`, s, s)
		result += "\n"
	}
	return result
}

func GetCSVFile(client *http.Client, msgChatID int64, sqID int) (*tgbotapi.DocumentConfig, error) {
	urlPath := fmt.Sprintf("http://%s:9999/api/v1/export/%d", epsURL, sqID)
	req, err := http.NewRequest(http.MethodGet, urlPath, nil)
	if err != nil {
		return nil, fmt.Errorf("GetCSVFile.NewRequest: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		//msg.Text = "We have problem to connect to our servers, Please try again later"
		return nil, fmt.Errorf("GetCSVFile.Get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GetCSVFile.StatusCode: unintented status code: %d", resp.StatusCode)
	}

	contentDisposition := resp.Header.Get("Content-Disposition")
	disposition, params, err := mime.ParseMediaType(contentDisposition)
	if err != nil {
		return nil, fmt.Errorf("GetCSVFile.ParseMediaType: %w", err)
	}
	fmt.Println(disposition)
	fileAbsPath := filepath.Join(
		"/tmp/eps/db/bot-storage",
		params["filename"],
	)

	// Create a new file to store the downloaded data
	file, err := os.Create(fileAbsPath)
	if err != nil {
		return nil, fmt.Errorf("GetCSVFile.CreateFile: %w", err)
	}
	defer file.Close()

	// Copy the response body to the file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return nil, fmt.Errorf("GetCSVFile.CopyFile: %w", err)
	}

	fmt.Println("File downloaded and stored successfully!")
	tgFile := tgbotapi.FilePath(fileAbsPath)
	msg := tgbotapi.NewDocument(msgChatID, tgFile)
	return &msg, nil
}

func CancelASearchQuery(client *http.Client, sqID int) (string, error) {
	reqBody, err := json.Marshal(map[string]int{
		"sq_id": sqID,
	})
	if err != nil {
		return "", fmt.Errorf("CancelASearchQuery.jsonMarshal: %w", err)
	}
	req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("http://%s:9999/api/v1/search", epsURL), bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("CancelASearchQuery.NewRequest: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		//msg.Text = "We have problem to connect to our servers, Please try again later"
		return "", fmt.Errorf("CancelASearchQuery.Get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Sprintf("There is no search query with `ID=%d`.", sqID), fmt.Errorf("CancelASearchQuery.StatusCode: unintented status code: %d", resp.StatusCode)
	}

	if resp.StatusCode != http.StatusNoContent {
		return "", fmt.Errorf("CancelASearchQuery.StatusCode: unintented status code: %d", resp.StatusCode)
	}
	return fmt.Sprintf("Search query with `ID=%d` canceled.", sqID), nil
}

func StartTheCrawler(client *http.Client) (_ string, err error) {
	if sq.Location == 0 {
		return "📍 Location is missing, please select your prefered location.", fmt.Errorf("bad search querey")
	}
	if sq.Language == "" {
		return "🌐 Language is missing, please select your prefered language.", fmt.Errorf("bad search querey")
	}
	if sq.Query == "" {
		return "Search Query is missing, please insert your prefered search query.", fmt.Errorf("bad search querey")
	}
	defer func() {
		sq = model.SearchQueryRequest{}
	}()
	reqBody, err := json.Marshal(sq)
	if err != nil {
		return "", fmt.Errorf("StartTheCrawler.jsonMarshal: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s:9999/api/v1/search", epsURL), bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("StartTheCrawler.NewRequest: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "Failed to communicate with crawler, please try again later.", fmt.Errorf("StartTheCrawler.Get: %w", err)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusConflict:
		return "Provided search items for starting of the crawler is already exist.\nPlease provide new one.", fmt.Errorf("conflict requerst")
	case http.StatusTooManyRequests:
		return "You have reached your searh rate limit..\nPlease Try after a while", fmt.Errorf("too many requests")
	case http.StatusOK:
		// Everything is OK
	default:
		logrus.Info("unkown error ", "status code: ", resp.StatusCode, "Query: ", sq)
		return "Internal server error,\n Please contact with support team.", fmt.Errorf("unknown error")
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("StartTheCrawler.ReadAll: %w", err)
	}
	var s StartCrawlerResponse
	err = json.Unmarshal(body, &s)
	if err != nil {
		return "", fmt.Errorf("StartTheCrawler.Unmarshal: %w", err)
	}
	return fmt.Sprintf("Crawler just started...\nYour Search Query ID is: %d", s.SQID), nil
}

func GetAllSearchQueries(client *http.Client, msgChatID int64, page int) ([]tgbotapi.MessageConfig, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s:9999/api/v1/search?page=%d", epsURL, page), nil)
	if err != nil {
		return nil, fmt.Errorf("GetAllSearchQueries.NewRequest: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GetAllSearchQueries.Get: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GetAllSearchQueries.StatusCode: unintented status code: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	var sqs model.SearchQueryResponse
	err = json.Unmarshal(bodyBytes, &sqs)
	if err != nil {
		return nil, fmt.Errorf("GetAllSearchQueries.Unmarshal: %w", err)
	}
	tgMessages := make([]tgbotapi.MessageConfig, 0, len(sqs.SQs))
	for _, sq := range sqs.SQs {
		msg := tgbotapi.NewMessage(msgChatID, "")
		msg.Text = fmt.Sprintf(
			`ID: %d
Query: %q
Location: %d
Language: %s
Created at: %s
`,
			sq.Id,
			sq.Query,
			sq.Location,
			strings.ToUpper(sq.Language),
			sq.CreatedAt.Local().String(),
		)
		tgMessages = append(tgMessages, msg)
	}
	if len(tgMessages) == 0 {
		return nil, sql.ErrNoRows
	}
	return tgMessages, nil
}

type StartCrawlerResponse struct {
	SQID int `json:"sq_id"` // SQID is Search Query ID
}

func isUserAuthenticated(userID string) bool {
	_, ok := authUsersMap[userID]
	return ok
}

func isUserAuthorized(userID string) bool {
	return authUsersMap[userID]
}

func authorizeUser(userID string) {
	authUsersMap[userID] = true
}

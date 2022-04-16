package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	strconv "strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dgraph-io/ristretto"
	"github.com/gookit/color"
	"github.com/kardianos/osext"
	"github.com/valyala/fasthttp"
)

type Settings struct {
	Tokens struct {
		Main string   `json:"main"`
		Alts []string `json:"alts"`
	} `json:"tokens"`
	Status struct {
		Main string `json:"main"`
		Alts string `json:"alts"`
	} `json:"status"`
	Nitro struct {
		Max        int  `json:"max"`
		Cooldown   int  `json:"cooldown"`
		MainSniper bool `json:"main_sniper"`
		Delay      bool `json:"delay"`
	} `json:"nitro"`
	Giveaway struct {
		Enable           bool     `json:"enable"`
		Delay            int      `json:"delay"`
		DM               string   `json:"dm"`
		DMDelay          int      `json:"dm_delay"`
		BlacklistWords   []string `json:"blacklist_words"`
		WhitelistWords   []string `json:"whitelist_words"`
		BlacklistServers []string `json:"blacklist_servers"`
	} `json:"giveaway"`
	Webhook struct {
		URL      string `json:"url"`
		GoodOnly bool   `json:"good_only"`
	} `json:"webhook"`
	BlacklistServers []string `json:"blacklist_servers"`
}

type Response struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

var (
	paymentSourceID string
	currentToken    string
	NitroSniped     int
	InviteSniped    int
	SniperRunning   bool
	InviteRunning   bool
	settings        Settings
	nbServers       int
	cache, _        = ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     1 << 30,
		BufferItems: 64,
	})
	reGiftLink        = regexp.MustCompile("(discord.com/gifts/|discordapp.com/gifts/|discord.gift/)([a-zA-Z0-9]+)")
	rePrivnote        = regexp.MustCompile("(https://privnote.com/[0-9A-Za-z]+)#([0-9A-Za-z]+)")
	rePrivnoteData    = regexp.MustCompile(`"data": "(.*)",`)
	reInviteServer    = regexp.MustCompile(`"name": "(.*)", "splash"`)
	reGiveaway        = regexp.MustCompile("You won the \\*\\*(.*)\\*\\*")
	reGiveawayMessage = regexp.MustCompile("<https://discordapp.com/channels/(.*)/(.*)/(.*)>")
	rePaymentSourceId = regexp.MustCompile(`("id": ")([0-9]+)"`)
	reNitroType       = regexp.MustCompile(` "name": "([ a-zA-Z]+)", "features"`)
)

func contains(array []string, value string) bool {
	for _, v := range array {
		if v == value {
			return true
		}
	}

	return false
}

func getPaymentSourceId() {
	var strRequestURI = []byte("https://discord.com/api/v8/users/@me/billing/payment-sources")
	req := fasthttp.AcquireRequest()
	req.Header.Set("authorization", settings.Tokens.Main)
	req.Header.SetMethodBytes([]byte("GET"))
	req.SetRequestURIBytes(strRequestURI)
	res := fasthttp.AcquireResponse()

	if err := fasthttp.Do(req, res); err != nil {
		return
	}

	fasthttp.ReleaseRequest(req)

	body := res.Body()

	id := rePaymentSourceId.FindStringSubmatch(string(body))

	if id == nil {
		paymentSourceID = "null"
	}
	if len(id) > 1 {
		paymentSourceID = id[2]
	}
}
func init() {
	executablePath, err := osext.ExecutableFolder()
	if err != nil {
		log.Fatal("Error: Couldn't determine working directory: " + err.Error())
	}
	os.Chdir(executablePath)
	file, err := ioutil.ReadFile("settings.json")
	if err != nil {
		fatalWithTime("[x] Failed read file: " + err.Error())
		time.Sleep(4 * time.Second)
		os.Exit(-1)
	}

	err = json.Unmarshal(file, &settings)
	if err != nil {
		fatalWithTime("[x] Failed to parse JSON file: " + err.Error())
		time.Sleep(4 * time.Second)
		os.Exit(-1)
	}

	NitroSniped = 0
	InviteSniped = 0
	SniperRunning = true
	InviteRunning = true
}
func timerEnd() {
	SniperRunning = true
	NitroSniped = 0
	logWithTime("<green>[<3] Hunting For Nitro </>")
}

func run(token string, finished chan bool, index int) {
	currentToken = token
	dg, err := discordgo.New(token)
	if err != nil {
		fatalWithTime("<red>[x] | Account may need to be verified | " + token + "," + err.Error())
	} else {
		err = dg.Open()
		if err != nil {
			logWithTime("<red>[x] | Account may have been terminated |" + token + "," + err.Error() + "</>")
		} else {
			nbServers += len(dg.State.Guilds)
			dg.AddHandler(messageCreate)
			if settings.Status.Alts != "" {
				_, _ = dg.UserUpdateStatus(discordgo.Status(settings.Status.Alts))
			}
		}
	}
	if index == len(settings.Tokens.Alts)-1 {
		finished <- true
	}
}

func deleteEmpty(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

func logWithTime(msg string) {
	timeStr := time.Now().Format("15:04:05")
	color.Println("<blue>" + timeStr + " </>" + msg)
}

func fatalWithTime(msg string) {
	timeStr := time.Now().Format("15:04:05 ")
	color.Println("<blue>" + timeStr + "</><red>" + msg + "</>")
	time.Sleep(4 * time.Second)
	os.Exit(-1)
}

func main() {

	if settings.Tokens.Main == "" {
		fatalWithTime("[x] You must put your token in settings.json")
	}

	finished := make(chan bool)

	settings.Tokens.Alts = deleteEmpty(settings.Tokens.Alts)

	if len(settings.Tokens.Alts) != 0 {
		for i, token := range settings.Tokens.Alts {
			go run(token, finished, i)
		}
	}

	var dg *discordgo.Session
	var err error

	if settings.Nitro.MainSniper {
		dg, err = discordgo.New(settings.Tokens.Main)

		if err != nil {
			fatalWithTime("[x] | Discord Account may need to be verified | " + settings.Tokens.Main + "," + err.Error())
			time.Sleep(4 * time.Second)
			os.Exit(-1)
		}

		err = dg.Open()
		if err != nil {
			fatalWithTime("[x] | Discord Account may have been terminated |" + settings.Tokens.Main + "," + err.Error())
			time.Sleep(5 * time.Second)
			os.Exit(-1)
		}

		dg.AddHandler(messageCreate)

		if settings.Status.Main != "" {
			_, _ = dg.UserUpdateStatus(discordgo.Status(settings.Status.Main))
		}

		nbServers += len(dg.State.Guilds)
	}

	if len(settings.Tokens.Alts) != 0 {
		<-finished
	}

	c := exec.Command("clear")
	c.Stdout = os.Stdout
	_ = c.Run()
	color.Blue.Println(`
       ..####...######..######..######..#####...######...........####...##..##..######..#####...######..#####..
       .##..##....##......##....##......##..##.....##...........##......###.##....##....##..##..##......##..##.
       .##..##....##......##....####....#####.....##.............####...##.###....##....#####...####....#####..
       .##..##....##......##....##......##..##...##.................##..##..##....##....##......##......##..##.
       ..####.....##......##....######..##..##..######...........####...##..##..######..##......######..##..##.
       ........................................................................................................
	                                             Vedeza Is hella sexy ;)
	   ........................................................................................................
`)

	getPaymentSourceId()

	color.Print("<cyan>Sniping Nitro</>")
	if settings.Nitro.MainSniper {
		color.Print("<cyan> for " + dg.State.User.Username + " on " + strconv.Itoa(nbServers) + " servers and " + strconv.Itoa(len(settings.Tokens.Alts)+1) + " alts </>\n\n")
	} else {
		color.Print("v on " + strconv.Itoa(nbServers) + " servers and " + strconv.Itoa(len(settings.Tokens.Alts)) + " alts </>\n\n")
	}
	logWithTime("<green>[<3] Sniper is ready</>")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	if settings.Nitro.MainSniper {
		_ = dg.Close()
	}
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if contains(settings.BlacklistServers, m.GuildID) {
		return
	}

	if reGiftLink.Match([]byte(m.Content)) && SniperRunning {
		checkGiftLink(s, m, m.Content, time.Now())
	} else if settings.Giveaway.Enable && !contains(settings.Giveaway.BlacklistServers, m.GuildID) && (strings.Contains(strings.ToLower(m.Content), "**giveaway**") || (strings.Contains(strings.ToLower(m.Content), "react with") && strings.Contains(strings.ToLower(m.Content), "giveaway"))) && m.Author.Bot {
		handleNewGiveaway(s, m)
	} else if (strings.Contains(strings.ToLower(m.Content), "giveaway") || strings.Contains(strings.ToLower(m.Content), "win") || strings.Contains(strings.ToLower(m.Content), "won")) && strings.Contains(m.Content, s.State.User.ID) && m.Author.Bot {
		handleGiveawayWon(s, m)
	}
}

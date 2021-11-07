package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/PuerkitoBio/goquery"
	"github.com/bwmarrin/discordgo"
)

// Variables used for command line parameters
var (
	Token string
)

var (
	ActiveSession *discordgo.Session
	ActiveMessage *discordgo.MessageCreate
	SpoilerCount  int    = 0
	MTGBoisRole   string = "<@&802701848619778098>"
)

func init() {

	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

func main() {

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	text := strings.Split(m.Content, " ")

	if text[0] != "!mtgSpoilers" {
		return
	}

	if len(text) == 1 {
		s.ChannelMessageSend(m.ChannelID, "Be sure to include some parameters. Start with '!mtgSpoilers help' if you are unsure what to do.")
		return
	}

	if text[1] == "help" {
		s.ChannelMessageSend(m.ChannelID, "Hello. I'm here to post new mtg spoilers directly to the discord channel. I'll let you know if I see something good!")
		return
	}

	if text[1] == "check" {
		ActiveMessage = m
		ActiveSession = s
		checkForNewSpoilers()
	}

}

func checkForNewSpoilers() string {
	scrapeMythicSpoiler()

	return ""
}

func scrapeMythicSpoiler() {
	response, err := http.Get("http://mythicspoiler.com/newspoilers.html")
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	// Create a goquery document from the HTTP response
	document, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		log.Fatal("Error loading HTTP response body. ", err)
	}

	// Find all links and process them with the function
	// defined earlier
	document.Find("img").Each(processElement)
}

// This will get called for each HTML element found
func processElement(index int, element *goquery.Selection) {
	// See if the src attribute exists on the element
	img, exists := element.Attr("src")
	if exists {
		if strings.Contains(img, "/cards/") {
			if SpoilerCount > 5 {
				return
			}
			imageSource := fmt.Sprintf("%s\nhttp://mythicspoiler.com/%s", MTGBoisRole, strings.Trim(img, "\n"))
			ActiveSession.ChannelMessageSend(ActiveMessage.ChannelID, imageSource)
			SpoilerCount++
		}
	}
}

// scraped HTML card examples

// <!--CARD CARD CARD CARD CARD CARD CARD--><div class="grid-card"><a href="
// mid/cards/nebelgastintruder.html
// "><img class="midcard" src="
// mid/cards/nebelgastintruder.jpg
// "></a><!--URL BELOW--><a href="
// https://twitter.com/coL_AliasV/status/1434911305379504129
// "><center><font face="'Arial Black', Gadget, sans-serif" color="#555555" size="-4">
// AliasV
// </font></a> </div><!--END CARD-->

//         <!--CARD CARD CARD CARD CARD CARD CARD--><div class="grid-card"><a href="
// mid/cards/covertcutpurse.html
// "><img class="midcard" src="
// mid/cards/covertcutpurse.jpg
// "></a><!--URL BELOW--><a href="
// https://twitter.com/val_pl_mafr/status/1434901009394974725?s=20
// "><center><font face="'Arial Black', Gadget, sans-serif" color="#555555" size="-4">
// @val_pl_mafr
// </font></a> </div><!--END CARD-->

//         <!--CARD CARD CARD CARD CARD CARD CARD--><div class="grid-card"><a href="
// mid/cards/covertcutpurse.html
// "><img class="" src="
// mid/cards/covetousgeist.jpg
// "></a><!--URL BELOW--><a href="
// https://twitter.com/val_pl_mafr/status/1434901009394974725?s=20
// "><center><font face="'Arial Black', Gadget, sans-serif" color="#555555" size="-4">
// @val_pl_mafr
// </font></a> </div><!--END CARD-->

//         <!--CARD CARD CARD CARD CARD CARD CARD--><div class="grid-card"><a href="
// mid/cards/burlybreaker.html
// "><img class="midcard" src="
// mid/cards/burlybreaker1.jpg
// "></a><!--URL BELOW--><a href="
// https://twitter.com/Blackneto_/status/1434895795967254532?s=20
// "><center><font face="'Arial Black', Gadget, sans-serif" color="#555555" size="-4">
// @Blackneto_
// </font></a> </div><!--END CARD-->

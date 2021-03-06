package backgroundcat

import (
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func OnMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	var link string
	for _, regex := range PasteRegexes {
		msgLink := regex.FindString(m.Content)
		if msgLink == "" {
			continue
		}
		link = pasteLinkToRaw(msgLink, regex)
		break
	}
	if link == "" {
		return
	}

	resp, err := http.Get(link)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	logs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	mistakes := AggregateMistakes(string(logs))
	if len(mistakes) == 0 {
		return
	}
	embed := discordgo.MessageEmbed{
		Title: "Automated Response:",
		Color: 0x11806A,
		Footer: &discordgo.MessageEmbedFooter{
			IconURL: "https://cdn.discordapp.com/emojis/280120125284417536.png?v=1",
			Text:    "This might not solve your problem, but it could be worth a try.",
		},
	}
	for _, mistake := range mistakes {
		embed.Fields = append(embed.Fields,
			&discordgo.MessageEmbedField{
				Name: string(mistake.severity), Value: mistake.message, Inline: true,
			})
	}
	if len(embed.Fields) == 0 {
		return
	}
	if _, err := s.ChannelMessageSendEmbed(m.ChannelID, &embed); err != nil {
		log.Println("Couldn't send message:", err)
	}
}

var (
	Pasteee  = regexp.MustCompile(`https?://paste\.ee/p/[^\s/]+`)
	Hastebin = regexp.MustCompile(`https?://has?tebin\.com/[^\s/]+`) // Also matches Hatebin
	Pastebin = regexp.MustCompile(`https?://pastebin\.com/[^\s/]+`)
	Pastegg  = regexp.MustCompile(`https?://paste\.gg/p/[^\s/]+/[^\s/]+`)
)

var PasteRegexes = [...]*regexp.Regexp{
	Pasteee,
	Hastebin,
	Pastebin,
	Pastegg,
}

func pasteLinkToRaw(link string, site *regexp.Regexp) string {
	switch site {
	case Pasteee:
		{
			return strings.Replace(link, "/p/", "/r/", 1)
		}
	case Hastebin, Pastebin:
		{
			return strings.Replace(link, ".com/", ".com/raw/", 1)
		}
	case Pastegg:
		{
			return link + "/raw"
		}
	}
	// We don't have a raw view, let's just return the link as-is
	return link
}

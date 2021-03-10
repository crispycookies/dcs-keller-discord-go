package dcskellerdiscordgo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sort"
	"strconv"
	"time"

	serverstatus "github.com/FleMo93/dcs-kellergeschwader-serverstatus-go"
	"github.com/bwmarrin/discordgo"
)

func verboseMsg(msg string, verbose bool) {
	if verbose {
		fmt.Println(msg)
	}
}

func readServerStatusFile(filePath string) (serverstatus.DCSServerStatus, error) {
	fileBytes, err := ioutil.ReadFile(filePath)
	status := serverstatus.DCSServerStatus{}
	if err != nil {
		return status, err
	}

	err = json.Unmarshal(fileBytes, &status)
	if err != nil {
		return status, err
	}

	return status, nil
}

func secondsToTimeString(time int) string {
	hours := fmt.Sprintf("%02s", strconv.Itoa(int(time)/60/60))
	minutes := fmt.Sprintf("%02s", strconv.Itoa(int(time)/60%60))
	seconds := fmt.Sprintf("%02s", strconv.Itoa(int(time)%60))
	return hours + ":" + minutes + ":" + seconds + " h"
}

func getPlayerListString(serverStatus serverstatus.DCSServerStatus) string {
	// plane - player group
	players := make(map[string][]serverstatus.DCSServerStatusPlayer)

	for _, player := range serverStatus.Players {
		players[player.Role] = append(players[player.Role], player)
	}

	var planeKeys []string
	for k := range players {
		planeKeys = append(planeKeys, k)
	}
	sort.Strings(planeKeys)

	listString := ""
	for _, planeName := range planeKeys {
		listString += "**" + planeName + "**\n"

		sort.Slice(players[planeName], func(i int, j int) bool {
			return players[planeName][i].Name < players[planeName][j].Name
		})

		for _, player := range players[planeName] {
			listString += "‏‏‎ ‎‏‏‎ ‎**`" + player.Name + "`** _" + secondsToTimeString(int(player.OnlineTime)) + "_\n"
		}
		listString += "\n"
	}

	return listString
}

// RunBot starts the dcs kellergeschwader discord bot
func RunBot(token string, botChannel string, serverStatusMessageID string, username string, password string, serverName string, serverStatusFile string, verbose bool) error {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return err
	}

	verboseMsg("Bot created", verbose)

	colorOnline := 3388721   //33b531
	colorOffline := 11878449 //b54031
	serverStatus, err := serverstatus.GetServerStatus(username, password, serverName)
	serverOnline := true

	if err != nil {
		if err.Error() == "Server not found" {
			serverOnline = false
		} else {
			return err
		}
	}
	verboseMsg("Returned server status", verbose)

	embedMessage := discordgo.MessageEmbed{}
	embedMessage.Title = "Server Status"
	embedMessage.Thumbnail = &discordgo.MessageEmbedThumbnail{
		URL: "https://upload.wikimedia.org/wikipedia/commons/thumb/f/ff/F16_drawing.svg/320px-F16_drawing.svg.png",
	}

	if serverOnline == true {
		embedMessage.Color = colorOnline
		embedMessage.Description += "**Online**\n"
		embedMessage.Description += "IP address: **" + serverStatus.IPADDRESS + ":" + serverStatus.PORT + "**\n"
		embedMessage.Description += "Mission: **" + serverStatus.MISSIONNAME + "**\n"

		if serverStatusFile != "" {
			status, err := readServerStatusFile(serverStatusFile)
			if err != nil {
				return err
			}

			embedMessage.Description += "Next mission: **" + secondsToTimeString(status.MissionTimeLeft) + "**\n"
			playerList := getPlayerListString(status)
			embedMessage.Description += "Players online: **" + strconv.Itoa(len(status.Players)) + "**"
			embedMessage.Description += "\n\n" + playerList
		} else {
			playersOnline, err := strconv.Atoi(serverStatus.PLAYERS)
			if err != nil {
				return err
			}
			playersOnline--
			embedMessage.Description += "Players online: **" + strconv.Itoa(playersOnline) + "**"
		}
	} else {
		embedMessage.Color = colorOffline
		embedMessage.Description += "**Offline**\n"
	}

	embedMessage.Timestamp = time.Now().UTC().Format("2006-01-02T15:04:05-0700")
	embedMessage.Footer = &discordgo.MessageEmbedFooter{
		Text: "Last update:",
	}

	message, err := session.ChannelMessageEditEmbed(botChannel, serverStatusMessageID, &embedMessage)
	if err != nil {
		return err
	}

	verboseMsg("Edited message", verbose)

	if message.Content != "" {
		verboseMsg("Clear message content", verbose)
		_, err := session.ChannelMessageEdit(message.ChannelID, message.ID, "")
		if err != nil {
			return err
		}
		verboseMsg("Message content cleared", verbose)
	}

	verboseMsg("Server status update finished", verbose)
	return nil
}

// CreateMessage Creates the messages needed for the bot status
func CreateMessage(token string, botChanngel string) ([]string, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return []string{}, err
	}

	msg, err := session.ChannelMessageSend(botChanngel, "Initial message")
	if err != nil {
		return []string{}, err
	}

	return []string{msg.ID}, nil
}

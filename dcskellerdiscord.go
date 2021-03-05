package dcskellerdiscordgo

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

type dcsServer struct {
	ID          string `json:"ID"`
	NAME        string `json:"NAME"`
	IPADDRESS   string `json:"IP_ADDRESS"`
	PORT        string `json:"PORT"`
	MISSIONNAME string `json:"MISSION_NAME"`
	MISSIONTIME string `json:"MISSION_TIME"`
	PLAYERS     string `json:"PLAYERS"`
	PLAYERSMAX  string `json:"PLAYERS_MAX"`
	PASSWORD    string `json:"PASSWORD"`
	URLTODETAIL string `json:"URL_TO_DETAIL"`
}

type dcsServerList struct {
	SERVERSMAXCOUNT int         `json:"SERVERS_MAX_COUNT"`
	SERVERSMAXDATE  string      `json:"SERVERS_MAX_DATE"`
	PLAYERSCOUNT    int         `json:"PLAYERS_COUNT"`
	MYSERVERS       []dcsServer `json:"MY_SERVERS"`
	SERVERS         []struct {
		NAME                 string `json:"NAME"`
		IPADDRESS            string `json:"IP_ADDRESS"`
		PORT                 string `json:"PORT"`
		MISSIONNAME          string `json:"MISSION_NAME"`
		MISSIONTIME          string `json:"MISSION_TIME"`
		PLAYERS              string `json:"PLAYERS"`
		PLAYERSMAX           string `json:"PLAYERS_MAX"`
		PASSWORD             string `json:"PASSWORD"`
		DESCRIPTION          string `json:"DESCRIPTION"`
		UALIAS0              string `json:"UALIAS_0"`
		MISSIONTIMEFORMATTED string `json:"MISSION_TIME_FORMATTED"`
	} `json:"SERVERS"`
}

func getServerStatus(username string, password string, serverName string) (dcsServer, error) {
	client := &http.Client{}

	url := "https://www.digitalcombatsimulator.com/en/personal/server/?ajax=y&_=" + strconv.FormatInt(time.Now().UTC().Unix(), 10)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return dcsServer{}, err
	}

	req.SetBasicAuth(username, password)
	resp, err := client.Do(req)
	if err != nil {
		return dcsServer{}, err
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	body := string(bodyBytes)
	jsonStart := strings.Index(body, "{")
	serverList := body[jsonStart:]
	serverStatus := &dcsServerList{}
	err = json.Unmarshal([]byte(serverList), serverStatus)
	if err != nil || body == "" {
		return dcsServer{}, err
	}

	for _, server := range serverStatus.MYSERVERS {
		if server.NAME == serverName {
			return server, nil
		}
	}
	return dcsServer{}, errors.New("Server not found")
}

// RunBot starts the dcs kellergeschwader discord bot
func RunBot(token string, botChannel string, serverStatusMessageID string, username string, password string, serverName string) error {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return err
	}

	colorOnline := 3388721   //33b531
	colorOffline := 11878449 //b54031
	serverStatus, err := getServerStatus(username, password, serverName)
	serverOnline := true

	if err != nil {
		if err.Error() == "Server not found" {
			serverOnline = false
		} else {
			return err
		}
	}

	embedMessage := discordgo.MessageEmbed{}
	embedMessage.Title = "Server Status"
	embedMessage.Thumbnail = &discordgo.MessageEmbedThumbnail{
		URL: "https://upload.wikimedia.org/wikipedia/commons/thumb/f/ff/F16_drawing.svg/320px-F16_drawing.svg.png",
	}

	if serverOnline == true {
		playersOnline, err := strconv.Atoi(serverStatus.PLAYERS)
		if err != nil {
			return err
		}
		playersOnline--

		embedMessage.Color = colorOnline
		embedMessage.Description += "**Online**\n"
		embedMessage.Description += "IP address: **" + serverStatus.IPADDRESS + ":" + serverStatus.PORT + "**\n"
		embedMessage.Description += "Mission: **" + serverStatus.MISSIONNAME + "**\n"
		embedMessage.Description += "Players online: **" + strconv.Itoa(playersOnline) + "**"
	} else {
		embedMessage.Color = colorOffline
		embedMessage.Description += "**Offline**\n"
	}

	embedMessage.Timestamp = time.Now().UTC().Format("2006-01-02T15:04:05-0700")
	embedMessage.Footer = &discordgo.MessageEmbedFooter{
		Text: "Last update",
	}

	session.ChannelMessageEditEmbed(botChannel, serverStatusMessageID, &embedMessage)
	return nil
}
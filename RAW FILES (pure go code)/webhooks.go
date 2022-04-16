package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/valyala/fasthttp"
)

func webhookNitro(code string, user *discordgo.User, guild string, channel string, status int, response string) {
	if settings.Webhook.URL == "" || (status <= 0 && settings.Webhook.GoodOnly) {
		return
	}
	var image = "https://cdn.discordapp.com/attachments/876887121024143370/898667054721286184/EThkxLwWsAMGQgp.png"
	var color = "7118070"

	if status == 0 {
		color = "7118070"
		image = ""
	} else if status == -1 {
		image = ""
		color = "7123370"
	}
	body := `
	{
	  "content": null,
	  "embeds": [
		{
		  "color": ` + color + `,
		  "fields": [
			{
			  "name": "Code",
			  "value": "` + code + `",
			  "inline": false
			},
			{
			  "name": "Guild",
			  "value": "` + guild + `",
			  "inline": true
			},
			{
			  "name": "Channel",
			  "value": "` + channel + `",
			  "inline": true
			},
			{
			  "name": "Response",
			  "value": "` + response + `",
			  "inline": false
			}
		  ],
		  "author": {
			"name": "Nitro Sniped !"
		  },
		  "footer": {
			"text": "OtterzSniper"
		  },
		  "thumbnail": {
			"url": "` + image + `"
		  }
		}
	  ],
	"username": "` + user.Username + `",
  	"avatar_url": "` + user.AvatarURL("") + `"
	}
	`

	req := fasthttp.AcquireRequest()
	req.Header.SetContentType("application/json")
	req.SetBody([]byte(body))
	req.Header.SetMethodBytes([]byte("POST"))
	req.SetRequestURIBytes([]byte(settings.Webhook.URL))
	res := fasthttp.AcquireResponse()

	if err := fasthttp.Do(req, res); err != nil {
		return
	}

	fasthttp.ReleaseRequest(req)
	fasthttp.ReleaseResponse(res)
}

func webhookGiveaway(prize string, user *discordgo.User, guild string, channel string) {
	if settings.Webhook.URL == "" {
		return
	}
	var color = "65290"

	if prize != "" {
		prize = `
			{
			  "name": "Prize",
			  "value": "` + prize + `",
			  "inline": false
			},`
	}

	body := `
	{
	  "content": null,
	  "embeds": [
		{
		  "color": ` + color + `,
		  "fields": [
			` + prize + `
			{
			  "name": "Guild",
			  "value": "` + guild + `",
			  "inline": true
			},
			{
			  "name": "Channel",
			  "value": "` + channel + `",
			  "inline": true
			}
		  ],
		  "author": {
			"name": "Giveaway Won !"
		  },
		  "footer": {
			"text": "OtterzSniper"
		  },
		  "thumbnail": {
        	"url": "https://cdn.discordapp.com/attachments/876864786195951652/898660975266369576/1hVdomV.png"
		  }
		}
	  ],
	"username": "` + user.Username + `",
  	"avatar_url": "` + user.AvatarURL("") + `"
	}
	`

	req := fasthttp.AcquireRequest()
	req.Header.SetContentType("application/json")
	req.SetBody([]byte(body))
	req.Header.SetMethodBytes([]byte("POST"))
	req.SetRequestURIBytes([]byte(settings.Webhook.URL))
	res := fasthttp.AcquireResponse()

	if err := fasthttp.Do(req, res); err != nil {
		return
	}

	fasthttp.ReleaseRequest(req)
	fasthttp.ReleaseResponse(res)
}

func webhookPrivnote(content string, user *discordgo.User, guild string, channel string, data string) {
	if settings.Webhook.URL == "" {
		return
	}
	var color = "65290"

	content = "`" + content + "`"
	data = "`" + data + "`"
	body := `
	{
	  "content": null,
	  "embeds": [
		{
		  "color": ` + color + `,
		  "fields": [
			{
			  "name": "Guild",
			  "value": "` + guild + `",
			  "inline": true
			},
			{
			  "name": "Channel",
			  "value": "` + channel + `",
			  "inline": true
			},
 			{
          	"name": "Content",
          	"value": "` + content + `"
        	},
			{
          	"name": "Encrypted",
          	"value": "` + data + `"
        	}
		  ],
		  "author": {
			"name": "Privnote Sniped !"
		  },
		  "footer": {
			"text": "OtterzSniper Made By Otterz"
		  },
		  "thumbnail": {
        	"url": "https://images.emojiterra.com/twitter/512px/1f4cb.png"
		  }
		}
	  ],
	"username": "` + user.Username + `",
  	"avatar_url": "` + user.AvatarURL("") + `"
	}
	`

	req := fasthttp.AcquireRequest()
	req.Header.SetContentType("application/json")
	req.SetBody([]byte(body))
	req.Header.SetMethodBytes([]byte("POST"))
	req.SetRequestURIBytes([]byte(settings.Webhook.URL))
	res := fasthttp.AcquireResponse()

	if err := fasthttp.Do(req, res); err != nil {
		return
	}

	fasthttp.ReleaseRequest(req)
	fasthttp.ReleaseResponse(res)
}

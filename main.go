package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/TheCreeper/go-notify"
	"github.com/google/go-github/v57/github"
)

func token() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	p := filepath.Join(home, ".config/github-desktop-notifications-token")
	b, err := os.ReadFile(p)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(b)), nil
}

func main() {
	tok, err := token()
	if err != nil {
		log.Fatal(err)
	}
	client := github.NewClient(nil).WithAuthToken(tok)

	idMapping := map[string]uint32{}

	for range time.Tick(time.Minute * 10) {
		ctx := context.Background()
		notifications, _, err := client.Activity.ListNotifications(ctx, &github.NotificationListOptions{All: true})
		if err != nil {
			log.Print(err)
			continue
		}

		for i := len(notifications) - 1; i >= 0; i-- {
			not := notifications[i]
			if *not.Unread {
				nt := githubToNotification(not)
				nid, _ := nt.Show()
				idMapping[*not.ID] = nid
			} else {
				if nid, ok := idMapping[*not.ID]; ok {
					notify.CloseNotification(nid)
					delete(idMapping, *not.ID)
				}
			}
		}
	}
}

func githubToNotification(gn *github.Notification) notify.Notification {
	summary := fmt.Sprintf("%s", *gn.Repository.FullName)
	body := fmt.Sprintf("%s: %s", *gn.Reason, *gn.Subject.Type)

	subjectID := ""
	switch *gn.Subject.Type {
	case "Issue":
		subjectID = fmt.Sprintf("#%s", path.Base(*gn.Subject.URL))
	case "PullRequest":
		subjectID = fmt.Sprintf("!%s", path.Base(*gn.Subject.URL))
	}

	switch *gn.Reason {
	case "comment":
		summary = fmt.Sprintf("%s %s", summary, subjectID)
		body = "💬 New Comment"
	case "mention":
		summary = fmt.Sprintf("%s %s", summary, subjectID)
		body = "💬 You were mentioned"
	case "subscribed":
		switch *gn.Subject.Type {
		case "PullRequest":
			summary = fmt.Sprintf("%s %s", summary, subjectID)
			body = fmt.Sprintf("🎁 %s", *gn.Subject.Title)
		case "Issue":
			summary = fmt.Sprintf("%s %s", summary, subjectID)
			body = fmt.Sprintf("🐞 %s", *gn.Subject.Title)
		case "Release":
			body = fmt.Sprintf("🚀 Released %s", *gn.Subject.Title)
		}
	}

	nt := notify.NewNotification(summary, body)
	nt.AppName = "GitHub"
	nt.Hints = map[string]any{
		notify.HintCategory: "github",
	}
	return nt
}

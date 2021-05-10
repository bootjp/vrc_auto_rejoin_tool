package vrcarjt

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Version struct {
	Major int
	Minor int
	Patch int
}

func getVersion(version string) (*Version, error) {
	if version == "" {
		return nil, errors.New("Please provide a valid version to parse")
	}
	nums := strings.Split(version, "")
	if len(nums) != 6 {
		return nil, fmt.Errorf("Please follow the proper versioning format, Got: %v, Expected: v0.0.0", version)
	}
	if nums[0] != "v" {
		return nil, fmt.Errorf("Please follow the proper versioning format, Got: %v, Expected: v0.0.0", version)
	}
	major, err := strconv.Atoi(nums[1])
	if err != nil {
		return nil, err
	}
	minor, err := strconv.Atoi(nums[3])
	if err != nil {
		return nil, err
	}
	patch, err := strconv.Atoi(nums[5])
	if err != nil {
		return nil, err
	}
	return &Version{
		Major: major,
		Minor: minor,
		Patch: patch,
	}, nil
}

type LatestVersion struct {
	URL       string `json:"url"`
	AssetsURL string `json:"assets_url"`
	UploadURL string `json:"upload_url"`
	HTMLURL   string `json:"html_url"`
	ID        int    `json:"id"`
	Author    struct {
		Login             string `json:"login"`
		ID                int    `json:"id"`
		NodeID            string `json:"node_id"`
		AvatarURL         string `json:"avatar_url"`
		GravatarID        string `json:"gravatar_id"`
		URL               string `json:"url"`
		HTMLURL           string `json:"html_url"`
		FollowersURL      string `json:"followers_url"`
		FollowingURL      string `json:"following_url"`
		GistsURL          string `json:"gists_url"`
		StarredURL        string `json:"starred_url"`
		SubscriptionsURL  string `json:"subscriptions_url"`
		OrganizationsURL  string `json:"organizations_url"`
		ReposURL          string `json:"repos_url"`
		EventsURL         string `json:"events_url"`
		ReceivedEventsURL string `json:"received_events_url"`
		Type              string `json:"type"`
		SiteAdmin         bool   `json:"site_admin"`
	} `json:"author"`
	NodeID          string    `json:"node_id"`
	TagName         string    `json:"tag_name"`
	TargetCommitish string    `json:"target_commitish"`
	Name            string    `json:"name"`
	Draft           bool      `json:"draft"`
	Prerelease      bool      `json:"prerelease"`
	CreatedAt       time.Time `json:"created_at"`
	PublishedAt     time.Time `json:"published_at"`
	Assets          []struct {
		URL      string `json:"url"`
		ID       int    `json:"id"`
		NodeID   string `json:"node_id"`
		Name     string `json:"name"`
		Label    string `json:"label"`
		Uploader struct {
			Login             string `json:"login"`
			ID                int    `json:"id"`
			NodeID            string `json:"node_id"`
			AvatarURL         string `json:"avatar_url"`
			GravatarID        string `json:"gravatar_id"`
			URL               string `json:"url"`
			HTMLURL           string `json:"html_url"`
			FollowersURL      string `json:"followers_url"`
			FollowingURL      string `json:"following_url"`
			GistsURL          string `json:"gists_url"`
			StarredURL        string `json:"starred_url"`
			SubscriptionsURL  string `json:"subscriptions_url"`
			OrganizationsURL  string `json:"organizations_url"`
			ReposURL          string `json:"repos_url"`
			EventsURL         string `json:"events_url"`
			ReceivedEventsURL string `json:"received_events_url"`
			Type              string `json:"type"`
			SiteAdmin         bool   `json:"site_admin"`
		} `json:"uploader"`
		ContentType        string    `json:"content_type"`
		State              string    `json:"state"`
		Size               int       `json:"size"`
		DownloadCount      int       `json:"download_count"`
		CreatedAt          time.Time `json:"created_at"`
		UpdatedAt          time.Time `json:"updated_at"`
		BrowserDownloadURL string    `json:"browser_download_url"`
	} `json:"assets"`
	TarballURL string `json:"tarball_url"`
	ZipballURL string `json:"zipball_url"`
	Body       string `json:"body"`
}

func (v *VRCAutoRejoinTool) GetLatestVersion() (*Version, error) {
	resp, err := http.Get("https://api.github.com/repos/bootjp/vrc_auto_rejoin_tool/releases/latest")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	var latestVersion LatestVersion
	err = json.Unmarshal(body, &latestVersion)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return getVersion(latestVersion.TagName)
}
func (v *VRCAutoRejoinTool) GetCurrentVersion() (*Version, error) {
	return getVersion(BuildVersion)
}

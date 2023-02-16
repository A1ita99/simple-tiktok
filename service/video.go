package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/charfole/simple-tiktok/config"
	"github.com/charfole/simple-tiktok/dao/mysql"
	"github.com/charfole/simple-tiktok/middleware"
	"github.com/charfole/simple-tiktok/model"
	ffmpeg "github.com/u2takey/ffmpeg-go"

	"github.com/tencentyun/cos-go-sdk-v5"
)

type FeedUser struct {
	ID             uint   `json:"id,omitempty"`
	Name           string `json:"name,omitempty"`
	FollowCount    uint   `json:"follow_count,omitempty"`
	FollowerCount  uint   `json:"follower_count,omitempty"`
	IsFollow       bool   `json:"is_follow,omitempty"`
	TotalFavorited uint   `json:"total_favorited"`
	FavoriteCount  uint   `json:"favorite_count"`
}

type FeedVideo struct {
	ID            uint     `json:"id,omitempty"`
	Author        FeedUser `json:"author,omitempty"`
	PlayURL       string   `json:"play_url,omitempty"`
	CoverURL      string   `json:"cover_url,omitempty"`
	FavoriteCount uint     `json:"favorite_count,omitempty"`
	CommentCount  uint     `json:"comment_count,omitempty"`
	IsFavorite    bool     `json:"is_favorite,omitempty"`
	Title         string   `json:"title,omitempty"`
}

// COSUpload 上传至云端，返回url
func COSUpload(fileName string, reader io.Reader) (string, error) {
	// bucketURL := fmt.Sprintf(objectstorage.COS_URL_FORMAT, objectstorage.COS_BUCKET_NAME, objectstorage.COS_APP_ID, objectstorage.COS_REGION)
	bucketURL := fmt.Sprintf(config.Info.COS.URLFormat, config.Info.COS.BucketName, config.Info.COS.AppID, config.Info.COS.Region)
	fmt.Println("bucketURL: ", bucketURL)
	u, _ := url.Parse(bucketURL)
	b := &cos.BaseURL{BucketURL: u}
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  config.Info.COS.SecretID,
			SecretKey: config.Info.COS.SecretKey,
		},
	})
	//path为本地的保存路径
	_, err := client.Object.Put(context.Background(), fileName, reader, nil)
	if err != nil {
		panic(err)
	}
	// return "https://charfolebase-1301984140.cos.ap-guangzhou.myqcloud.com/" + fileName, nil
	return bucketURL + "/" + fileName, nil
}

// ExampleReadFrameAsJpeg 获取封面
func ExampleReadFrameAsJpeg(inFileName string, frameNum int) io.Reader {
	buf := bytes.NewBuffer(nil)
	err := ffmpeg.Input(inFileName).
		Filter("select", ffmpeg.Args{fmt.Sprintf("gte(n,%d)", frameNum)}).
		Output("pipe:", ffmpeg.KwArgs{"vframes": 1, "format": "image2", "vcodec": "mjpeg"}).
		WithOutput(buf, os.Stdout).
		Run()
	if err != nil {
		panic(err)
	}
	return buf
}

func PackFeedResponse(strToken string, videoList []model.Video) (feedVideoList []FeedVideo, newTime int64) {
	var haveToken bool
	if strToken == "" {
		haveToken = false
	} else {
		haveToken = true
	}

	for _, video := range videoList {
		var tmp FeedVideo
		var user model.User
		tmp.ID = video.ID
		// get the author info
		err := mysql.GetAUserByID(video.AuthorID, &user)
		var feedUser FeedUser
		if err == nil { // author exists
			feedUser.ID = user.ID
			feedUser.FollowerCount = user.FollowerCount
			feedUser.FollowCount = user.FollowCount
			feedUser.Name = user.Name
			feedUser.TotalFavorited = user.TotalFavorited
			feedUser.FavoriteCount = user.FavoriteCount
			feedUser.IsFollow = false
			if haveToken {
				// check the token
				tokenStruct, ok := middleware.CheckToken(strToken)
				// check the user follows the author or not
				if ok && time.Now().Unix() <= tokenStruct.ExpiresAt {
					var uid1 = tokenStruct.UserID // user id
					var uid2 = video.AuthorID     // author id
					// if current user is the author, do not show the follow button
					if uid1 == uid2 {
						feedUser.IsFollow = true
					} else {
						feedUser.IsFollow = IsFollowing(uid1, uid2)
					}
				}
			}
		}
		tmp.PlayURL = video.PlayURL
		tmp.Author = feedUser
		tmp.CommentCount = video.CommentCount
		tmp.CoverURL = video.CoverURL
		tmp.FavoriteCount = video.FavoriteCount
		tmp.IsFavorite = false
		if haveToken {
			// check the token
			tokenStruct, ok := middleware.CheckToken(strToken)
			if ok && time.Now().Unix() <= tokenStruct.ExpiresAt {
				var uid = tokenStruct.UserID // user id
				var vid = video.ID           // video id
				if mysql.IsFavorite(uid, vid) {
					tmp.IsFavorite = true
				}
			}
		}
		tmp.Title = video.Title
		feedVideoList = append(feedVideoList, tmp)
		// next query time is the oldest time in current videolist
		newTime = video.CreatedAt.Unix()
	}
	return
}

// Get the video list
func FeedGet(lastTime int64) ([]model.Video, error) {
	// max num of videos
	const videoNum = 2
	// reset the latest time
	if lastTime == 0 {
		lastTime = time.Now().Unix()
	}
	// print the time for test
	strTime := fmt.Sprint(time.Unix(lastTime, 0).Format("2006-01-02 15:04:05"))
	fmt.Println("query time: ", strTime)

	var videoList []model.Video
	// err := mysql.DB.Table("videos").Where("created_at < ?", strTime).Order("created_at desc").Limit(videoNum).Find(&videoList).Error
	err := mysql.GetVideoByTime(strTime, videoNum, &videoList)

	return videoList, err
}

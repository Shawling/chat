package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path"
)

//当获取不到头像时，返回 ErrNoAvatarURL 错误
var ErrNoAvatarURL = errors.New("Chat: Unalbe to get an avatar URL.")

type Avatar interface {
	GetAvatarURL(c *client) (string, error)
}

type TryAvatars []Avatar

var UseTryAvatars Avatar = TryAvatars{
	UseFileSystemAvatar,
	UseGravatar,
	UseAuthAvatar,
}

func (a TryAvatars) GetAvatarURL(c *client) (string, error) {
	for _, mehtod := range a {
		if url, err := mehtod.GetAvatarURL(c); err == nil {
			return url, nil
		}
	}
	return "", ErrNoAvatarURL
}

type AuthAvatar struct{}

var UseAuthAvatar AuthAvatar

func (AuthAvatar) GetAvatarURL(c *client) (string, error) {
	if url, ok := c.userData["avatar_url"]; ok {
		if urlStr, ok := url.(string); ok {
			return urlStr, nil
		}
	}
	return "", ErrNoAvatarURL
}

type GravatarAvatar struct{}

var UseGravatar GravatarAvatar

func (GravatarAvatar) GetAvatarURL(c *client) (string, error) {
	if userID, ok := c.userData["userid"]; ok {
		if userIDStr, ok := userID.(string); ok {
			return fmt.Sprintf("//www.gravatar.com/avatar/%s", userIDStr), nil
		}
	}
	return "", ErrNoAvatarURL
}

type FileSystemAvatar struct{}

var UseFileSystemAvatar FileSystemAvatar

func (FileSystemAvatar) GetAvatarURL(c *client) (string, error) {
	if userid, ok := c.userData["userid"]; ok {
		if useridStr, ok := userid.(string); ok {
			files, err := ioutil.ReadDir("avatars")
			if err != nil {
				return "", ErrNoAvatarURL
			}
			for _, file := range files {
				if file.IsDir() {
					continue
				}
				if match, _ := path.Match(useridStr+"*", file.Name()); match {
					return "/avatars/" + file.Name(), nil
				}
			}
		}
	}
	return "", ErrNoAvatarURL
}

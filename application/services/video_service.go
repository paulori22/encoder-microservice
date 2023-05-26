package services

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"cloud.google.com/go/storage"
	"github.com/paulori22/encoder-microservice/application/repositories"
	"github.com/paulori22/encoder-microservice/domain"
)

type VideoService struct {
	Video           *domain.Video
	VideoRepository repositories.VideoRepository
}

func NewVideoService() VideoService {
	return VideoService{}
}

func localStoragePath() string {
	return os.Getenv("localStoragePath")
}

func (v *VideoService) VideoFolderPath() string {
	return localStoragePath() + "/" + v.Video.ID
}

func (v *VideoService) VideoPath() string {
	return v.VideoFolderPath() + ".mp4"
}

func (v *VideoService) VideoFragPath() string {
	return v.VideoFolderPath() + ".frag"
}

func (v *VideoService) Download(bucketName string) error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)

	if err != nil {
		return err
	}

	bkt := client.Bucket(bucketName)
	obj := bkt.Object(v.Video.FilePath)

	r, err := obj.NewReader(ctx)

	if err != nil {
		return err
	}

	defer r.Close()

	body, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	f, err := os.Create(v.VideoPath())
	if err != nil {
		return err
	}

	_, err = f.Write(body)
	if err != nil {
		return err
	}

	defer f.Close()

	log.Printf("video %v has been stored", v.Video.ID)

	return nil
}

func (v *VideoService) Fragment() error {

	err := os.Mkdir(v.VideoFolderPath(), os.ModePerm)
	if err != nil {
		return err
	}

	source := v.VideoPath()
	target := v.VideoFragPath()

	cmd := exec.Command("mp4fragment", source, target)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	printOutput(output)

	return nil
}

func (v *VideoService) Encode() error {

	cmdArgs := []string{}
	cmdArgs = append(cmdArgs, v.VideoFragPath())
	cmdArgs = append(cmdArgs, "--use-segment-timeline")
	cmdArgs = append(cmdArgs, "-o")
	cmdArgs = append(cmdArgs, v.VideoFolderPath())
	cmdArgs = append(cmdArgs, "-f")
	cmdArgs = append(cmdArgs, "--exec-dir")
	cmdArgs = append(cmdArgs, "/opt/bento4/bin/")
	cmd := exec.Command("mp4dash", cmdArgs...)

	output, err := cmd.CombinedOutput()

	if err != nil {
		return err
	}

	printOutput(output)

	return nil
}

func (v *VideoService) Finish() error {
	err := os.Remove(v.VideoPath())
	if err != nil {
		log.Printf("error removing mp4 file (%v)\n", v.VideoPath())
		return err
	}

	err = os.Remove(v.VideoFragPath())
	if err != nil {
		log.Printf("error removing frag file (%v)\n", v.VideoFragPath())
		return err
	}

	err = os.RemoveAll(v.VideoFolderPath())
	if err != nil {
		log.Printf("error removing folder (%v)\n", v.VideoFolderPath())
		return err
	}

	log.Println("temp files have been removed: ", v.Video.ID)

	return nil
}

func (v *VideoService) InsertVideo() error {

	_, err := v.VideoRepository.Insert(v.Video)

	if err != nil {
		return err
	}

	return nil
}

func printOutput(out []byte) {
	if len(out) > 0 {
		log.Printf("======> Output: %s\n", string(out))
	}
}

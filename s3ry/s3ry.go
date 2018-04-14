package s3ry

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type s3ry struct {
	sess *session.Session
	svc  *s3.S3
}

/*

s3ry

*/

func NewS3ry() *s3ry {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("ap-northeast-1")},
	))

	svc := s3.New(sess)
	s := &s3ry{
		sess: sess,
		svc:  svc,
	}
	return s
}

func (s s3ry) SelectOperation() string {
	items := []PromptItems{
		{Key: 0, Val: "ダウンロード"},
		{Key: 1, Val: "アップロード"},
		{Key: 2, Val: "オブジェクトリスト"},
	}
	result := run("何をしますか？", items)
	return result
}

func (s s3ry) ListBuckets() string {
	sps("バケットの検索中です...")
	input := &s3.ListBucketsInput{}
	listBuckets, err := s.svc.ListBuckets(input)
	if err != nil {
		awsErrorPrint(err)
	}
	items := []PromptItems{}
	for key, val := range listBuckets.Buckets {
		items = append(items, PromptItems{Key: key, Val: *val.Name, Tag: "Bucket"})
	}
	spe()
	result := run("どのバケットを利用しますか?", items)
	return result
}

func (s s3ry) ListObjects(bucket string) string {
	sps("オブジェクトの検索中です...")
	items := []PromptItems{}
	// @todo- start 共通化
	key := 0
	err := s.svc.ListObjectsPages(&s3.ListObjectsInput{Bucket: aws.String(bucket)},
		func(listObjects *s3.ListObjectsOutput, lastPage bool) bool {
			for _, item := range listObjects.Contents {
				if strings.HasSuffix(*item.Key, "/") == false {
					items = append(items, PromptItems{Key: key, Val: *item.Key, LastModified: *item.LastModified, Tag: "Object"})
					key++
				}
			}
			return !lastPage
		})
	if err != nil {
		awsErrorPrint(err)
	}
	// @todo 並び替えオプションをフラグをグローバルにもたせて、KeyでもSort出来るようにする
	sort.Slice(items, func(i, j int) bool {
		return items[i].LastModified.After(items[j].LastModified)
	})
	// @todo-end

	spe()

	fmt.Println("オブジェクト数：", len(items))
	result := run("どのファイルを取得しますか?", items)
	return result
}

func (s s3ry) GetObject(bucket string, objectKey string) {
	sps("オブジェクトのダウンロード中です...")
	filename := filepath.Base(objectKey)
	file, err := os.Create(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "")
		os.Exit(1)
	}
	inputGet := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectKey),
	}
	downloader := s3manager.NewDownloader(s.sess)
	result, err := downloader.Download(file, inputGet)
	// @todo fileのクローズしていない！
	if err != nil {
		// @todo file消したほうが良い
		awsErrorPrint(err)
	}
	spe()
	fmt.Printf("ファイルをダウンロードしました, %s, %d bytes\n", filename, result)
}

func (s s3ry) UploadObject(bucket string) {
	dir := dirwalk()
	items := []PromptItems{}
	for key, val := range dir {
		// @todo Bucketってなんだ
		items = append(items, PromptItems{Key: key, Val: val, Tag: "Bucket"})
	}
	selected := run("どのファイルをアップロードしますか?", items)

	sps("オブジェクトのアップロード中です...")
	uploadObject := selected
	uploader := s3manager.NewUploader(s.sess)
	f, err := os.Open(uploadObject)
	if err != nil {
		fmt.Fprintf(os.Stderr, "")
		os.Exit(1)
	}

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(uploadObject),
		Body:   f,
	})
	if err != nil {
		// @todo fileをクローズするか否かを検討
		awsErrorPrint(err)
	}
	spe()
	fmt.Printf("ファイルをアップロードしました, %s \n", uploadObject)
}

func (s s3ry) SaveObjectList(bucket string) {
	sps("オブジェクトの検索中です...")
	// @todo- start 共通化
	items := []PromptItems{}
	key := 0
	err := s.svc.ListObjectsPages(&s3.ListObjectsInput{Bucket: aws.String(bucket)},
		func(listObjects *s3.ListObjectsOutput, lastPage bool) bool {
			for _, item := range listObjects.Contents {
				if strings.HasSuffix(*item.Key, "/") == false {
					items = append(items, PromptItems{Key: key, Val: *item.Key, Size: *item.Size, LastModified: *item.LastModified, Tag: "Object"})
					key++
				}
			}
			return !lastPage
		})
	if err != nil {
		awsErrorPrint(err)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].LastModified.After(items[j].LastModified)
	})
	if err != nil {
		awsErrorPrint(err)
	}
	// @todo-end
	spe()
	t := time.Now()
	// @todo strftimeに変更した方が良い https://github.com/lestrrat-go/strftime
	ObjectListFileName := "ObjectList-" + t.Format("2006-01-02-15-04-05") + ".txt"
	file, err := os.Create(ObjectListFileName)
	for _, item := range items {
		_, err = file.Write(([]byte)("./" + item.Val + "," + strconv.FormatInt(item.Size, 10) + "\n"))
		if err != nil {
			fmt.Println("書き込みエラーです", err)
			os.Exit(1)
		}
	}
	// @todo fileのクローズしていない！
	fmt.Println("オブジェクトリストを作成しました:" + ObjectListFileName)
	os.Exit(0)
}

package service

import (
	foundationr2 "foundation/foundation-r2"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	cfr2 "meta/cf-r2"
	metaerror "meta/meta-error"
	metahttp "meta/meta-http"
	"meta/set"
	"meta/singleton"
	"path"
	"regexp"
	"strings"
	"web/config"
)

type R2ImageService struct {
}

var singletonR2ImageService = singleton.Singleton[R2ImageService]{}

func GetR2ImageService() *R2ImageService {
	return singletonR2ImageService.GetInstance(
		func() *R2ImageService {
			return &R2ImageService{}
		},
	)
}

func (s *R2ImageService) ProcessContentFromMarkdown(
	content string,
	oldContent *string,
	prefix string,
	newPrefix string,
) (
	string,
	[]*foundationr2.R2ImageUrl,
	error,
) {
	bucketName := "didapipa-oj"
	r2Client := cfr2.GetSubsystem().GetClient(bucketName)
	if r2Client == nil {
		return content, nil, metaerror.New("R2 client is not available")
	}

	r2Url := config.GetConfig().R2Url

	reg := regexp.MustCompile(`!\[[^\]]*\]\(([^)]+)\)`)
	newMatches := reg.FindAllStringSubmatch(content, -1)

	var needDeleteUrls []*s3.ObjectIdentifier

	if oldContent != nil {
		oldMatches := reg.FindAllStringSubmatch(*oldContent, -1)

		oldImageUrls := set.New[string]()
		r2ImagePrefix := metahttp.UrlJoin(r2Url, prefix)
		for _, match := range oldMatches {
			if len(match) > 1 {
				imageURL := match[1]
				if strings.HasPrefix(imageURL, r2ImagePrefix) {
					oldImageUrls.Add(imageURL)
				}
			}
		}
		newImageUrls := set.New[string]()
		for _, match := range newMatches {
			if len(match) > 1 {
				imageURL := match[1]
				if strings.HasPrefix(imageURL, r2ImagePrefix) {
					newImageUrls.Add(imageURL)
				}
			}
		}
		oldImageUrls.Foreach(
			func(oldUrl *string) bool {
				if !newImageUrls.Contains(*oldUrl) {
					needDeleteUrls = append(
						needDeleteUrls, &s3.ObjectIdentifier{
							Key: aws.String(strings.TrimPrefix(strings.TrimPrefix(*oldUrl, r2Url), "/")),
						},
					)
				}
				return true
			},
		)
	}
	prefixUpdating := metahttp.UrlJoin(r2Url, "uploading", prefix)
	// 判断是否存在需要迁移的临时图片
	var needUpdateUrls []*foundationr2.R2ImageUrl
	for _, match := range newMatches {
		if len(match) > 1 {
			imageURL := match[1]
			if strings.HasPrefix(imageURL, prefixUpdating) {
				fileName := path.Base(imageURL)
				newUrl := metahttp.UrlJoin(r2Url, newPrefix, fileName)
				needUpdateUrls = append(
					needUpdateUrls, &foundationr2.R2ImageUrl{
						Old: imageURL,
						New: newUrl,
					},
				)
			}
		}
	}
	//把所有的needUpdateUrls替换为新的路径
	if len(needUpdateUrls) > 0 {
		for _, url := range needUpdateUrls {
			content = strings.ReplaceAll(content, url.Old, url.New)
		}
	}
	if len(needDeleteUrls) > 0 {
		// 删除不再使用的图片
		_, err := r2Client.DeleteObjects(
			&s3.DeleteObjectsInput{
				Bucket: aws.String(bucketName),
				Delete: &s3.Delete{
					Objects: needDeleteUrls,
					Quiet:   aws.Bool(true),
				},
			},
		)
		if err != nil {
			return "", nil, metaerror.Wrap(err, "failed to delete old images")
		}
	}
	return content, needUpdateUrls, nil
}

func (s *R2ImageService) MoveImageAfterSave(needUpdateUrls []*foundationr2.R2ImageUrl) error {

	bucketName := "didapipa-oj"
	r2Client := cfr2.GetSubsystem().GetClient(bucketName)
	if r2Client == nil {
		return metaerror.New("R2 client is not available")
	}

	r2Url := config.GetConfig().R2Url

	var finalErr error

	// 把所有的needUpdateUrls移动到新的路径
	for _, url := range needUpdateUrls {
		oldKey := strings.TrimPrefix(strings.TrimPrefix(url.Old, r2Url), "/")
		newKey := strings.TrimPrefix(strings.TrimPrefix(url.New, r2Url), "/")
		// 生成预签名链接
		_, err := r2Client.CopyObject(
			&s3.CopyObjectInput{
				Bucket:     aws.String(bucketName),
				CopySource: aws.String(path.Join(bucketName, oldKey)),
				Key:        aws.String(newKey),
			},
		)
		finalErr = metaerror.Join(finalErr, err)
	}
	return finalErr
}

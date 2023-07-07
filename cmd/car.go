package cmd

import (
	"fmt"
	"github.com/alanshaw/go-carbites"
	"github.com/cheggaaa/pb/v3"
	chainstoragesdk "github.com/paradeum-team/chainstorage-sdk"
	sdkcode "github.com/paradeum-team/chainstorage-sdk/code"
	"github.com/paradeum-team/chainstorage-sdk/consts"
	"github.com/paradeum-team/chainstorage-sdk/model"
	"github.com/paradeum-team/chainstorage-sdk/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/ulule/deepcopier"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

func init() {
	//carUploadCmd.Flags().StringP("Bucket", "b", "", "桶名称")
	//carUploadCmd.Flags().StringP("Object", "o", "", "上传对象路径")
	//
	//carImportCmd.Flags().StringP("Bucket", "b", "", "桶名称")
	//carImportCmd.Flags().StringP("Carfile", "c", "", "car文件标识")

	//objectRenameCmd.Flags().StringP("Bucket", "b", "", "桶名称")
	//objectRenameCmd.Flags().StringP("Object", "o", "", "对象名称")
	//objectRenameCmd.Flags().StringP("Cid", "c", "", "Cid")
	//objectRenameCmd.Flags().StringP("Rename", "r", "", "重命名")
	//objectRenameCmd.Flags().BoolP("Force", "f", false, "有冲突的时候强制覆盖")
	//
	//objectRemoveCmd.Flags().StringP("Bucket", "b", "", "桶名称")
	//objectRemoveCmd.Flags().StringP("Object", "o", "", "对象名称")
	//objectRemoveCmd.Flags().StringP("Cid", "c", "", "Cid")
	//objectRemoveCmd.Flags().BoolP("Force", "f", false, "有冲突的时候强制覆盖")
}

// region CAR Upload

//var carUploadCmd = &cobra.Command{
//	Use:     "put",
//	Short:   "put",
//	Long:    "upload object",
//	Example: "gcscmd put FILE[/DIR...] cs://BUCKET",
//
//	Run: func(cmd *cobra.Command, args []string) {
//		carUploadRun(cmd, args)
//	},
//}

func carUploadRun(cmd *cobra.Command, args []string) {
	// 桶名称
	bucketName := GetBucketName(args)
	if err := checkBucketName(bucketName); err != nil {
		Error(cmd, args, err)
	}

	// 上传对象路径
	dataPath := GetDataPath(args)

	//// 上传 carfile
	//carFile, err := cmd.Flags().GetString("carFile")
	//if err != nil {
	//	Error(cmd, args, err)
	//}

	sdk, err := chainstoragesdk.New(&appConfig)
	if err != nil {
		Error(cmd, args, err)
	}

	// 确认桶数据有效性
	respBucket, err := sdk.Bucket.GetBucketByName(bucketName)
	if err != nil {
		Error(cmd, args, err)
	}

	code := int(respBucket.Code)
	if code != http.StatusOK {
		Error(cmd, args, errors.New(respBucket.Msg))
	}

	// 桶ID
	bucketId := respBucket.Data.Id

	// 检查上传数据使用限制
	storageNetworkCode := respBucket.Data.StorageNetworkCode
	err = checkDataUsageLimitation(sdk, storageNetworkCode, dataPath)
	if err != nil {
		Error(cmd, args, err)
	}

	// 对象上传
	response, err := UploadData(sdk, bucketId, dataPath)
	if err != nil {
		Error(cmd, args, err)
	}

	carUploadRunOutput(cmd, args, response)
}

func carUploadRunOutput(cmd *cobra.Command, args []string, resp model.ObjectCreateResponse) {
	respCode := int(resp.Code)

	if respCode != http.StatusOK {
		Error(cmd, args, errors.New(resp.Msg))
	}

	carUploadOutput := CarUploadOutput{
		RequestId: resp.RequestId,
		Code:      resp.Code,
		Msg:       resp.Msg,
		Status:    resp.Status,
	}

	err := deepcopier.Copy(&resp.Data).To(&carUploadOutput.Data)
	if err != nil {
		Error(cmd, args, err)
	}

	templateContent := `
CID: {{.Data.ObjectCid}}
Name: {{.Data.ObjectName}}
`

	t, err := template.New("carUploadTemplate").Parse(templateContent)
	if err != nil {
		Error(cmd, args, err)
	}

	err = t.Execute(os.Stdout, carUploadOutput)
	if err != nil {
		Error(cmd, args, err)
	}
}

type CarUploadOutput struct {
	RequestId string       `json:"requestId,omitempty"`
	Code      int32        `json:"code,omitempty"`
	Msg       string       `json:"msg,omitempty"`
	Status    string       `json:"status,omitempty"`
	Data      ObjectOutput `json:"objectOutput,omitempty"`
}

type CarUploadResponse struct {
	RequestId string      `json:"requestId,omitempty"`
	Code      int32       `json:"code,omitempty"`
	Msg       string      `json:"msg,omitempty"`
	Status    string      `json:"status,omitempty"`
	Data      interface{} `json:"data,omitempty"`
}

// 上传数据
func UploadData(sdk *chainstoragesdk.CssClient, bucketId int, dataPath string) (model.ObjectCreateResponse, error) {
	//response := model.CarResponse{}
	response := model.ObjectCreateResponse{}

	// 数据路径为空
	if len(dataPath) == 0 {
		return response, sdkcode.ErrCarUploadFileInvalidDataPath
	}

	// 数据路径无效
	fileInfo, err := os.Stat(dataPath)
	if os.IsNotExist(err) {
		return response, sdkcode.ErrCarUploadFileInvalidDataPath
	} else if err != nil {
		//log.Errorf("Fail to get file path, error:%+v\n", err)
		log.WithError(err).WithField("dataPath", dataPath).Error("fail to return stat of file")
		return response, err
	}

	carFileUploadReq := model.CarFileUploadReq{}
	// 上传为目录的情况
	if fileInfo.IsDir() {
		notEmpty, err := isFolderNotEmpty(dataPath)
		if err != nil {
			log.WithError(err).WithField("dataPath", dataPath).Error("fail to check uploadiong folder")
			return response, err
		}

		if !notEmpty {
			return response, sdkcode.ErrCarUploadFileInvalidDataFolder
		}

		//if fileInfo.IsDir() || isCarFile {
		carFileUploadReq.ObjectTypeCode = consts.ObjectTypeCodeDir
	}

	//if !isCarFile {
	fileDestination := sdk.Car.GenerateTempFileName(utils.CurrentDate()+"_", ".tmp")
	//fileDestination := GenerateTempFileName("", ".tmp")
	carVersion := sdkConfig.CarVersion
	//log.Infof("UploadData carVersion:%d, fileDestination:%s, dataPath:%s\n", carVersion, fileDestination, dataPath)

	log.WithFields(logrus.Fields{
		"fileDestination": fileDestination,
		"dataPath":        dataPath,
		"carVersion":      carVersion,
		"begintime":       GetTimestampString(),
	}).Info("Create car file start")
	//fmt.Printf("Create car file start, begintime:%s\n", GetTimestampString())
	// 创建Car文件
	err = sdk.Car.CreateCarFile(dataPath, fileDestination)
	if err != nil {
		log.WithError(err).
			WithFields(logrus.Fields{
				"fileDestination": fileDestination,
				"dataPath":        dataPath,
			}).Error("Fail to create car file")
		//log.Errorf("Fail to create car file, error:%+v\n", err)
		return response, sdkcode.ErrCarUploadFileCreateCarFileFail
	}
	//fmt.Printf("Create car file finish, endtime:%s\n", GetTimestampString())
	log.WithFields(logrus.Fields{
		"fileDestination": fileDestination,
		"dataPath":        dataPath,
		"carVersion":      carVersion,
		"endtime":         GetTimestampString(),
	}).Info("Create car file finish")

	defer func(fileDestination string) {
		//if !viper.GetBool("cli.cleanTmpData") {
		if !cliConfig.CleanTmpData {
			return
		}

		err := os.Remove(fileDestination)
		if err != nil {
			log.WithError(err).
				WithFields(logrus.Fields{
					"fileDestination": fileDestination,
				}).Error("Fail to remove car file")
			//fmt.Printf("Error:%+v\n", err)
		}
	}(fileDestination)
	//}

	// 解析CAR文件，获取DAG信息，获取文件或目录的CID
	rootLink := model.RootLink{}
	err = sdk.Car.ParseCarFile(fileDestination, &rootLink)
	if err != nil {
		log.WithError(err).
			WithFields(logrus.Fields{
				"fileDestination": fileDestination,
			}).Error("Fail to parse car file")
		//log.Errorf("Fail to parse car file, error:%+v\n", err)
		return response, sdkcode.ErrCarUploadFileParseCarFileFail
	}

	rootCid := rootLink.RootCid.String()
	objectCid := rootLink.Cid.String()
	objectSize := int64(rootLink.Size)
	objectName := rootLink.Name

	//if isCarFile {
	//	objectCid = rootCid
	//	objectSize = fileInfo.Size()
	//
	//	filename := filepath.Base(dataPath)
	//	filename = strings.TrimSuffix(filename, ".car")
	//	objectName = filename
	//}

	// 设置请求参数
	carFileUploadReq.BucketId = bucketId
	carFileUploadReq.ObjectCid = objectCid
	carFileUploadReq.ObjectSize = objectSize
	carFileUploadReq.ObjectName = objectName
	carFileUploadReq.FileDestination = fileDestination
	carFileUploadReq.CarFileCid = rootCid

	// 计算文件sha256
	sha256, err := utils.GetFileSha256ByPath(fileDestination)
	if err != nil {
		log.WithError(err).
			WithFields(logrus.Fields{
				"fileDestination": fileDestination,
			}).Error("Fail to calculate file sha256")
		//log.Errorf("Fail to calculate file sha256, error:%+v\n", err)
		return response, sdkcode.ErrCarUploadFileComputeCarFileHashFail
	}
	carFileUploadReq.RawSha256 = sha256

	// 使用Root CID秒传检查
	objectExistResponse, err := sdk.Object.IsExistObjectByCid(objectCid)
	if err != nil {
		log.WithError(err).
			WithFields(logrus.Fields{
				"objectCid":           objectCid,
				"objectExistResponse": objectExistResponse,
			}).Error("Fail to check if exist object")
		//log.Errorf("Fail to check if exist object, error:%+v\n", err)
		return response, sdkcode.ErrCarUploadFileReferenceObjcetFail
	}

	// CID存在，执行秒传操作
	objectExistCheck := objectExistResponse.Data
	if objectExistCheck.IsExist {
		response, err := sdk.Car.ReferenceObject(&carFileUploadReq)
		if err != nil {
			log.WithError(err).
				WithFields(logrus.Fields{
					"carFileUploadReq": carFileUploadReq,
					"response":         response,
				}).Error("Fail to reference object")
			//fmt.Printf("Error:%+v\n", err)
			return response, sdkcode.ErrCarUploadFileReferenceObjcetFail
		}

		return response, nil
	}

	// CAR文件大小，超过分片阈值
	carFileInfo, err := os.Stat(fileDestination)
	if err != nil {
		log.WithError(err).WithField("fileDestination", fileDestination).Error("fail to return stat of file")
		return response, err
	}

	carFileSize := carFileInfo.Size()
	carFileShardingThreshold := sdk.Config.CarFileShardingThreshold

	// 生成CAR分片文件上传
	if carFileSize > int64(carFileShardingThreshold) {
		response, err = UploadBigCarFile(sdk, &carFileUploadReq)
		if err != nil {
			log.WithError(err).
				WithFields(logrus.Fields{
					"carFileUploadReq": carFileUploadReq,
					"response":         response,
				}).Error("Fail to upload big car file")
			return response, sdkcode.ErrCarUploadFileFail
		}

		return response, nil
	}

	// 普通上传
	file, err := os.Open(fileDestination)
	if err != nil {
		log.WithError(err).WithField("fileDestination", fileDestination).Error("fail to return stat of file")
		return response, err
	}
	defer file.Close()

	bar := pb.Start64(carFileSize).SetWriter(os.Stdout).Set(pb.Bytes, true)
	bar.SetRefreshRate(100 * time.Millisecond)
	defer bar.Finish()

	extReader := bar.NewProxyReader(file)
	defer extReader.Close()

	log.WithFields(logrus.Fields{
		"carFileUploadReq": carFileUploadReq,
		"begintime":        GetTimestampString(),
	}).Info("Upload car file start")
	//fmt.Printf("UploadData => UploadCarFileExt, parameter carFileUploadReq:%+v\n", carFileUploadReq)
	//fmt.Printf("Upload car file start, begintime:%s\n", GetTimestampString())
	response, err = sdk.Car.UploadCarFileExt(&carFileUploadReq, extReader)
	if err != nil {
		log.WithError(err).
			WithFields(logrus.Fields{
				"carFileUploadReq": carFileUploadReq,
				"response":         response,
			}).Error("Fail to upload car file")
		return response, sdkcode.ErrCarUploadFileFail
	}
	//fmt.Printf("Upload car file finish, endtime:%s\n", GetTimestampString())
	//fmt.Printf("UploadData => UploadCarFileExt, response:%+v\n", response)
	log.WithFields(logrus.Fields{
		"response": response,
		"endtime":  GetTimestampString(),
	}).Info("Upload car file finish")

	return response, err
}

// 上传大CAR文件
func UploadBigCarFile(sdk *chainstoragesdk.CssClient, req *model.CarFileUploadReq) (model.ObjectCreateResponse, error) {
	response := model.ObjectCreateResponse{}

	log.WithFields(logrus.Fields{
		"req":       req,
		"begintime": GetTimestampString(),
	}).Info("Generate sharding car files start")
	//fmt.Printf("UploadBigCarFile => GenerateShardingCarFiles, parameter req:%+v\n", req)
	//fmt.Printf("Generate sharding car files start, begintime:%s\n", GetTimestampString())
	// 生成CAR分片文件
	shardingCarFileUploadReqs := []model.CarFileUploadReq{}
	//err := sdk.Car.GenerateShardingCarFiles(req, &shardingCarFileUploadReqs)
	err := GenerateShardingCarFiles(req, &shardingCarFileUploadReqs)
	if err != nil {
		log.WithError(err).
			WithFields(logrus.Fields{
				"req":                       req,
				"shardingCarFileUploadReqs": shardingCarFileUploadReqs,
			}).Error("Fail to generate sharding car files")
		return response, err
	}
	//fmt.Printf("Generate sharding car files finish, endtime:%s\n", GetTimestampString())
	//fmt.Printf("UploadBigCarFile => GenerateShardingCarFiles, parameter shardingCarFileUploadReqs:%+v\n", shardingCarFileUploadReqs)
	log.WithFields(logrus.Fields{
		"shardingCarFileUploadReqs": shardingCarFileUploadReqs,
		"endtime":                   GetTimestampString(),
	}).Info("Generate sharding car files finish")

	// 删除CAR分片文件
	defer func(shardingCarFileUploadReqs []model.CarFileUploadReq) {
		//if !viper.GetBool("cli.cleanTmpData") {
		if !cliConfig.CleanTmpData {
			return
		}

		for i := range shardingCarFileUploadReqs {
			fileDestination := shardingCarFileUploadReqs[i].FileDestination
			err := os.Remove(fileDestination)
			if err != nil {
				log.WithError(err).
					WithFields(logrus.Fields{
						"fileDestination": fileDestination,
					}).Error("Fail to remove sharding car file")
				//fmt.Printf("Error:%+v\n", err)
				//logger.Errorf("file.Delete %s err: %v", fileDestination, err)
			}
		}
	}(shardingCarFileUploadReqs)

	// 计算总文件大小
	totalSize := int64(0)
	for i, _ := range shardingCarFileUploadReqs {
		totalSize += shardingCarFileUploadReqs[i].ObjectSize
	}

	bar := pb.Start64(totalSize).SetWriter(os.Stdout).Set(pb.Bytes, true)
	bar.SetRefreshRate(100 * time.Millisecond)
	defer bar.Finish()

	// 上传CAR文件分片
	//uploadingReqs := []model.CarFileUploadReq{}
	//deepcopier.Copy(&shardingCarFileUploadReqs).To(&uploadingReqs)
	maxRetries := 3
	retryDelay := time.Duration(3) * time.Second

	log.WithFields(logrus.Fields{
		"shardingCarFileUploadReqs": shardingCarFileUploadReqs,
		"begintime":                 GetTimestampString(),
	}).Info("Upload sharding car files start")
	//fmt.Printf("Upload sharding car file start, begintime:%s\n", GetTimestampString())
	uploadRespList := []model.ShardingCarFileUploadResponse{}
	for i, _ := range shardingCarFileUploadReqs {
		for j := 0; j < maxRetries; j++ {
			uploadingReq := model.CarFileUploadReq{}
			deepcopier.Copy(&shardingCarFileUploadReqs[i]).To(&uploadingReq)

			file, err := os.Open(uploadingReq.FileDestination)
			defer file.Close()
			//fi, err := file.Stat()
			//size := fi.Size()
			extReader := bar.NewProxyReader(file)
			defer extReader.Close()

			log.WithFields(logrus.Fields{
				"uploadingReq": uploadingReq,
				"index":        i,
				"retry":        j,
			}).Info("upload sharding car file")
			//fmt.Printf("UploadBigCarFile => UploadShardingCarFileExt, index:%d, parameter uploadingReq:%+v\n", i, uploadingReq)
			uploadResp, err := sdk.Car.UploadShardingCarFileExt(&uploadingReq, extReader)
			if err == nil && uploadResp.Code == http.StatusOK {
				uploadRespList = append(uploadRespList, uploadResp)
				break
			}

			// 记录日志
			//fmt.Printf("UploadBigCarFile => UploadShardingCarFileExt, index:%d, uploadResp:%+v\n", i, uploadResp)
			log.WithError(err).
				WithFields(logrus.Fields{
					"uploadingReq": uploadingReq,
					"uploadResp":   uploadResp,
					"index":        i,
					"retry":        j,
				}).Error("Fail to upload sharding car file")

			if j == maxRetries-1 {
				// 尝试maxRetries次失败
				if err != nil {
					return response, err
				} else if uploadResp.Code != http.StatusOK {
					return response, errors.New(response.Msg)
				}
			}

			time.Sleep(retryDelay)
		}
	}
	//fmt.Printf("Upload sharding car file finish, endtime:%s\n", GetTimestampString())
	log.WithFields(logrus.Fields{
		"shardingCarFileUploadReqs": shardingCarFileUploadReqs,
		"endtime":                   GetTimestampString(),
	}).Info("Upload sharding car files finish")

	log.WithFields(logrus.Fields{
		"req":       req,
		"begintime": GetTimestampString(),
	}).Info("Confirm sharding car files start")
	//fmt.Printf("UploadBigCarFile => ConfirmShardingCarFiles, parameter req:%+v\n", req)
	//fmt.Printf("Confirm sharding car file start, begintime:%s\n", GetTimestampString())
	// 确认分片上传成功
	response, err = sdk.Car.ConfirmShardingCarFiles(req)
	if err != nil {
		log.WithError(err).
			WithFields(logrus.Fields{
				"req":      req,
				"response": response,
			}).Error("Fail to Confirm sharding car files")
		return response, err
	}
	//fmt.Printf("Confirm sharding car file finish, endtime:%s\n", GetTimestampString())
	//fmt.Printf("UploadBigCarFile => ConfirmShardingCarFiles, response:%+v\n", response)
	log.WithFields(logrus.Fields{
		"response": response,
		"endtime":  GetTimestampString(),
	}).Info("Confirm sharding car files finish")

	return response, nil
}

// endregion CAR Upload

// region CAR Import

//var carImportCmd = &cobra.Command{
//	Use:     "import",
//	Short:   "import",
//	Long:    "import car file",
//	Example: "gcscmd import  ./aaa.car cs://BUCKET",
//
//	Run: func(cmd *cobra.Command, args []string) {
//		//cmd.Help()
//		//fmt.Printf("%s %s\n", cmd.Name(), strconv.Itoa(offset))
//		carImportRun(cmd, args)
//	},
//}

func carImportRun(cmd *cobra.Command, args []string) {
	// 桶名称
	bucketName := GetBucketName(args)
	if err := checkBucketName(bucketName); err != nil {
		Error(cmd, args, err)
	}

	// 上传对象路径
	dataPath := GetDataPath(args)

	// CAR文件类型检查
	if !strings.HasSuffix(strings.ToLower(dataPath), ".car") {
		//err := sdkcode.ErrCarUploadFileInvalidDataPath
		//Error(cmd, args, err)
		Error(cmd, args, errors.New("please specify car format file with .car suffix"))
	}

	//// CAR文件标识
	//carFile, err := cmd.Flags().GetString("carFile")
	//if err != nil {
	//	Error(cmd, args, err)
	//}

	sdk, err := chainstoragesdk.New(&appConfig)
	if err != nil {
		Error(cmd, args, err)
	}

	// 确认桶数据有效性
	respBucket, err := sdk.Bucket.GetBucketByName(bucketName)
	if err != nil {
		Error(cmd, args, err)
	}

	code := int(respBucket.Code)
	if code != http.StatusOK {
		Error(cmd, args, errors.New(respBucket.Msg))
	}

	// 桶ID
	bucketId := respBucket.Data.Id

	// 检查上传数据使用限制
	storageNetworkCode := respBucket.Data.StorageNetworkCode
	err = checkDataUsageLimitation(sdk, storageNetworkCode, dataPath)
	if err != nil {
		Error(cmd, args, err)
	}

	// 对象上传
	response, err := ImportData(sdk, bucketId, dataPath)
	if err != nil {
		Error(cmd, args, err)
	}

	carImportRunOutput(cmd, args, response)
}

// 检查上传数据使用限制
func checkDataUsageLimitation(sdk *chainstoragesdk.CssClient, storageNetworkCode int, dataPath string) error {
	// 检查可用空间
	usersQuotaResp, err := sdk.Bucket.GetUsersQuotaByStorageNetworkCode(storageNetworkCode)
	if err != nil {
		return err
	}

	usersQuota := usersQuotaResp.Data
	usersQuotaDetails := usersQuota.Details
	if len(usersQuotaDetails) == 0 {
		return sdkcode.ErrBucketQuotaFetchFail
	}

	// 基础版本
	isBasicVersion := usersQuota.PackagePlanId == 21001
	availableStorageSpace := int64(0)
	availableFileAmount := int64(0)
	availableUploadDirItems := int64(0)

	for _, usersQuotaDetail := range usersQuotaDetails {
		// 空间存储限制
		if usersQuotaDetail.ConstraintName == consts.ConstraintStorageSpace.String() {
			//availableStorageSpace = usersQuotaDetail.Available
			availableStorageSpace = usersQuotaDetail.LimitedQuota - usersQuotaDetail.UsedQuota
		}

		// 对象存储限制
		if usersQuotaDetail.ConstraintName == consts.ConstraintFileLimited.String() {
			//availableFileAmount = usersQuotaDetail.Available
			availableFileAmount = usersQuotaDetail.LimitedQuota - usersQuotaDetail.UsedQuota
		}

		// 上传文件夹条目限制
		if usersQuotaDetail.ConstraintName == consts.ConstraintUploadDirItems.String() {
			//availableUploadDirItems = usersQuotaDetail.Available
			availableUploadDirItems = usersQuotaDetail.LimitedQuota - usersQuotaDetail.UsedQuota
		}
	}

	// 可用文件存储限制超限
	if availableFileAmount <= 0 {
		return sdkcode.ErrCarUploadFileExccedObjectAmountUsage
	}

	// 获取上传数据使用量
	fileAmount, totalSize, err := getUploadingDataUsage(dataPath)
	if err != nil {
		return err
	}

	// 上传文件夹条目限制超限
	if fileAmount > availableUploadDirItems {
		return sdkcode.ErrCarUploadFileExccedObjectAmountUsage
	}

	if isBasicVersion {
		// 可用存储空间超限
		if totalSize > availableStorageSpace {
			return sdkcode.ErrCarUploadFileExccedStorageSpaceUsage
		}
	}

	return nil
}

func carImportRunOutput(cmd *cobra.Command, args []string, resp model.ObjectCreateResponse) {
	code := resp.Code
	if code != http.StatusOK {
		Error(cmd, args, errors.New(resp.Msg))
	}

	carImportOutput := CarImportOutput{
		RequestId: resp.RequestId,
		Code:      resp.Code,
		Msg:       resp.Msg,
		Status:    resp.Status,
	}

	err := deepcopier.Copy(&resp.Data).To(&carImportOutput.Data)
	if err != nil {
		Error(cmd, args, err)
	}

	//	导入 car 文件
	//	通过命令向固定桶内导入 car 文件对象
	//
	//	模版
	//
	//	gcscmd import  ./aaa.car cs://BUCKET
	//	BUCKET
	//
	//	桶名称
	//
	//	carfile
	//
	//	car文件标识
	//
	//	命令行例子
	//
	//	当前目录
	//
	//	gcscmd import ./aaa.car cs://bbb
	//	绝对路径
	//
	//	gcscmd import /home/pz/aaa.car cs://bbb
	//	相对路径
	//
	//	gcscmd import ../pz/aaa.car cs://bbb
	//	响应
	//
	//	过程
	//
	//	################                                                                15%
	//		QmWgnG7pPjG31w328hZyALQ2BgW5aQrZyKpT47jVpn8CNo        Tarkov.mp4
	//	完成
	//
	//CID:    QmWgnG7pPjG31w328hZyALQ2BgW5aQrZyKpT47jVpn8CNo
	//Name:Tarkov.mp4
	//	报错
	//
	//Error: This is not a carfile

	templateContent := `
CID: {{.Data.ObjectCid}}
Name: {{.Data.ObjectName}}
`

	t, err := template.New("carImportTemplate").Parse(templateContent)
	if err != nil {
		Error(cmd, args, err)
	}

	err = t.Execute(os.Stdout, carImportOutput)
	if err != nil {
		Error(cmd, args, err)
	}
}

// 导入CAR文件数据
func ImportData(sdk *chainstoragesdk.CssClient, bucketId int, dataPath string) (model.ObjectCreateResponse, error) {
	//response := model.CarResponse{}
	response := model.ObjectCreateResponse{}

	// 数据路径为空
	if len(dataPath) == 0 {
		return response, sdkcode.ErrCarUploadFileInvalidDataPath
	}

	// 数据路径无效
	fileInfo, err := os.Stat(dataPath)
	if os.IsNotExist(err) {
		return response, sdkcode.ErrCarUploadFileInvalidDataPath
	} else if err != nil {
		log.WithError(err).WithField("dataPath", dataPath).Error("fail to return stat of file")
		return response, err
	}

	fileDestination := dataPath

	// 解析CAR文件，获取DAG信息，获取文件或目录的CID
	rootLink := model.RootLink{}
	err = sdk.Car.ParseCarFile(fileDestination, &rootLink)
	if err != nil {
		//fmt.Printf("Error:%+v\n", err)
		log.WithError(err).
			WithFields(logrus.Fields{
				"fileDestination": fileDestination,
			}).Error("Fail to parse car file")
		return response, sdkcode.ErrCarUploadFileParseCarFileFail
	}

	rootCid := rootLink.RootCid.String()
	//objectCid := rootLink.Cid.String()
	//objectSize := int64(rootLink.Size)
	//objectName := rootLink.Name

	objectCid := rootCid
	objectSize := fileInfo.Size()

	filename := filepath.Base(dataPath)
	filename = strings.TrimSuffix(filename, ".car")
	objectName := filename

	// 设置请求参数
	carFileUploadReq := model.CarFileUploadReq{}
	carFileUploadReq.BucketId = bucketId
	carFileUploadReq.ObjectCid = objectCid
	carFileUploadReq.ObjectSize = objectSize
	carFileUploadReq.ObjectName = objectName
	carFileUploadReq.FileDestination = fileDestination
	carFileUploadReq.CarFileCid = rootCid

	// 导入CAR文件时，对象类型为目录
	carFileUploadReq.ObjectTypeCode = consts.ObjectTypeCodeDir

	// 计算文件sha256
	sha256, err := utils.GetFileSha256ByPath(fileDestination)
	if err != nil {
		//fmt.Printf("Error:%+v\n", err)
		log.WithError(err).
			WithFields(logrus.Fields{
				"fileDestination": fileDestination,
			}).Error("Fail to calculate file sha256")
		return response, sdkcode.ErrCarUploadFileComputeCarFileHashFail
	}
	carFileUploadReq.RawSha256 = sha256

	// 使用Root CID秒传检查
	objectExistResponse, err := sdk.Object.IsExistObjectByCid(objectCid)
	if err != nil {
		//fmt.Printf("Error:%+v\n", err)
		log.WithError(err).
			WithFields(logrus.Fields{
				"objectCid":           objectCid,
				"objectExistResponse": objectExistResponse,
			}).Error("Fail to check if exist object")
		return response, sdkcode.ErrCarUploadFileReferenceObjcetFail
	}

	// CID存在，执行秒传操作
	objectExistCheck := objectExistResponse.Data
	if objectExistCheck.IsExist {
		response, err := sdk.Car.ReferenceObject(&carFileUploadReq)
		if err != nil {
			//fmt.Printf("Error:%+v\n", err)
			log.WithError(err).
				WithFields(logrus.Fields{
					"carFileUploadReq": carFileUploadReq,
					"response":         response,
				}).Error("Fail to reference object")
			return response, sdkcode.ErrCarUploadFileReferenceObjcetFail
		}

		return response, err
	}

	// CAR文件大小，超过分片阈值
	carFileSize := fileInfo.Size()
	carFileShardingThreshold := sdk.Config.CarFileShardingThreshold

	// 生成CAR分片文件上传
	if carFileSize > int64(carFileShardingThreshold) {
		response, err = ImportBigCarFile(sdk, &carFileUploadReq)
		if err != nil {
			log.WithError(err).
				WithFields(logrus.Fields{
					"carFileUploadReq": carFileUploadReq,
					"response":         response,
				}).Error("Fail to import big car file")
			return response, sdkcode.ErrCarUploadFileFail
		}

		return response, nil
	}

	// 普通上传
	file, err := os.Open(fileDestination)
	if err != nil {
		log.WithError(err).WithField("fileDestination", fileDestination).Error("fail to return stat of file")
		return response, err
	}
	defer file.Close()

	bar := pb.Start64(carFileSize).SetWriter(os.Stdout).Set(pb.Bytes, true)
	bar.SetRefreshRate(100 * time.Millisecond)
	defer bar.Finish()

	extReader := bar.NewProxyReader(file)
	defer extReader.Close()

	log.WithFields(logrus.Fields{
		"carFileUploadReq": carFileUploadReq,
		"begintime":        GetTimestampString(),
	}).Info("Import car file start")
	//fmt.Printf("ImportData => ImportCarFileExt, parameter carFileUploadReq:%+v\n", carFileUploadReq)
	//fmt.Printf("Import car file start, begintime:%s\n", GetTimestampString())
	response, err = sdk.Car.ImportCarFileExt(&carFileUploadReq, extReader)
	if err != nil {
		log.WithError(err).
			WithFields(logrus.Fields{
				"carFileUploadReq": carFileUploadReq,
				"response":         response,
			}).Error("Fail to import car file")
		return response, sdkcode.ErrCarUploadFileFail
	}
	//fmt.Printf("Import car file finish, endtime:%s\n", GetTimestampString())
	//fmt.Printf("ImportData => ImportCarFileExt, response:%+v\n", response)
	log.WithFields(logrus.Fields{
		"response": response,
		"endtime":  GetTimestampString(),
	}).Info("Import car file finish")

	return response, err
}

// 导入大CAR文件
func ImportBigCarFile(sdk *chainstoragesdk.CssClient, req *model.CarFileUploadReq) (model.ObjectCreateResponse, error) {
	response := model.ObjectCreateResponse{}

	log.WithFields(logrus.Fields{
		"req":       req,
		"begintime": GetTimestampString(),
	}).Info("Generate sharding car files start")
	//fmt.Printf("ImportBigCarFile => GenerateShardingCarFiles, parameter req:%+v\n", req)
	//fmt.Printf("Generate sharding car files start, begintime:%s\n", GetTimestampString())
	// 生成CAR分片文件
	shardingCarFileUploadReqs := []model.CarFileUploadReq{}
	//err := sdk.Car.GenerateShardingCarFiles(req, &shardingCarFileUploadReqs)
	err := GenerateShardingCarFiles(req, &shardingCarFileUploadReqs)
	if err != nil {
		log.WithError(err).
			WithFields(logrus.Fields{
				"req":                       req,
				"shardingCarFileUploadReqs": shardingCarFileUploadReqs,
			}).Error("Fail to generate sharding car files")
		return response, err
	}
	//fmt.Printf("Generate sharding car files finish, endtime:%s\n", GetTimestampString())
	//fmt.Printf("ImportBigCarFile => GenerateShardingCarFiles, parameter shardingCarFileUploadReqs:%+v\n", shardingCarFileUploadReqs)
	log.WithFields(logrus.Fields{
		"shardingCarFileUploadReqs": shardingCarFileUploadReqs,
		"endtime":                   GetTimestampString(),
	}).Info("Generate sharding car files finish")

	// 删除CAR分片文件
	defer func(shardingCarFileUploadReqs []model.CarFileUploadReq) {
		//if !viper.GetBool("cli.cleanTmpData") {
		if !cliConfig.CleanTmpData {
			return
		}

		for i := range shardingCarFileUploadReqs {
			fileDestination := shardingCarFileUploadReqs[i].FileDestination
			err := os.Remove(fileDestination)
			if err != nil {
				//fmt.Printf("Error:%+v\n", err)
				log.WithError(err).
					WithFields(logrus.Fields{
						"fileDestination": fileDestination,
					}).Error("Fail to remove sharding car file")
			}
		}
	}(shardingCarFileUploadReqs)

	totalSize := int64(0)
	for i, _ := range shardingCarFileUploadReqs {
		totalSize += shardingCarFileUploadReqs[i].ObjectSize
	}

	bar := pb.Start64(totalSize).SetWriter(os.Stdout).Set(pb.Bytes, true)
	bar.SetRefreshRate(100 * time.Millisecond)
	defer bar.Finish()

	// 上传CAR文件分片
	//uploadingReqs := []model.CarFileUploadReq{}
	//deepcopier.Copy(&shardingCarFileUploadReqs).To(&uploadingReqs)
	maxRetries := cliConfig.MaxRetries
	retryDelay := time.Duration(cliConfig.RetryDelay) * time.Second

	log.WithFields(logrus.Fields{
		"shardingCarFileUploadReqs": shardingCarFileUploadReqs,
		"begintime":                 GetTimestampString(),
	}).Info("Upload sharding car files start")
	//fmt.Printf("Upload sharding car file start, begintime:%s\n", GetTimestampString())
	uploadRespList := []model.ShardingCarFileUploadResponse{}
	for i, _ := range shardingCarFileUploadReqs {
		for j := 0; j < maxRetries; j++ {
			uploadingReq := model.CarFileUploadReq{}
			deepcopier.Copy(&shardingCarFileUploadReqs[i]).To(&uploadingReq)

			file, err := os.Open(uploadingReq.FileDestination)
			defer file.Close()
			//fi, err := file.Stat()
			//size := fi.Size()
			extReader := bar.NewProxyReader(file)
			defer extReader.Close()

			log.WithFields(logrus.Fields{
				"uploadingReq": uploadingReq,
				"index":        i,
				"retry":        j,
			}).Info("import sharding car file")
			//fmt.Printf("ImportBigCarFile => ImportShardingCarFileExt, index:%d, parameter uploadingReq:%+v\n", i, uploadingReq)
			uploadResp, err := sdk.Car.ImportShardingCarFileExt(&uploadingReq, extReader)
			if err == nil && uploadResp.Code == http.StatusOK {
				uploadRespList = append(uploadRespList, uploadResp)
				break
			}

			// 记录日志
			//fmt.Printf("ImportBigCarFile => ImportShardingCarFileExt, index:%d, uploadResp:%+v\n", i, uploadResp)
			log.WithError(err).
				WithFields(logrus.Fields{
					"uploadingReq": uploadingReq,
					"uploadResp":   uploadResp,
					"index":        i,
					"retry":        j,
				}).Error("Fail to import sharding car file")

			if j == maxRetries-1 {
				// 尝试maxRetries次失败
				if err != nil {
					return response, err
				} else if uploadResp.Code != http.StatusOK {
					return response, errors.New(response.Msg)
				}
			}

			time.Sleep(retryDelay)
		}
	}
	//fmt.Printf("Upload sharding car file finish, endtime:%s\n", GetTimestampString())
	log.WithFields(logrus.Fields{
		"shardingCarFileUploadReqs": shardingCarFileUploadReqs,
		"endtime":                   GetTimestampString(),
	}).Info("Upload sharding car files finish")

	log.WithFields(logrus.Fields{
		"req":       req,
		"begintime": GetTimestampString(),
	}).Info("Confirm sharding car files start")
	//fmt.Printf("ImportBigCarFile => ConfirmShardingCarFiles, parameter req:%+v\n", req)
	//fmt.Printf("Confirm sharding car file start, begintime:%s\n", GetTimestampString())
	// 确认分片上传成功
	response, err = sdk.Car.ConfirmShardingCarFiles(req)
	if err != nil {
		log.WithError(err).
			WithFields(logrus.Fields{
				"req":      req,
				"response": response,
			}).Error("Fail to Confirm sharding car files")
		return response, err
	}
	//fmt.Printf("Confirm sharding car file finish, endtime:%s\n", GetTimestampString())
	//fmt.Printf("ImportBigCarFile => ConfirmShardingCarFiles, response:%+v\n", response)
	log.WithFields(logrus.Fields{
		"response": response,
		"endtime":  GetTimestampString(),
	}).Info("Confirm sharding car files finish")
	return response, nil
}

type CarImportOutput struct {
	RequestId string       `json:"requestId,omitempty"`
	Code      int32        `json:"code,omitempty"`
	Msg       string       `json:"msg,omitempty"`
	Status    string       `json:"status,omitempty"`
	Data      ObjectOutput `json:"objectOutput,omitempty"`
}

//type ObjectOutput struct {
//	Id             int       `json:"id" comment:"对象ID"`
//	BucketId       int       `json:"bucketId" comment:"桶主键"`
//	ObjectName     string    `json:"objectName" comment:"对象名称（255字限制）"`
//	ObjectTypeCode int       `json:"objectTypeCode" comment:"对象类型编码"`
//	ObjectSize     int64     `json:"objectSize" comment:"对象大小（字节）"`
//	IsMarked       int       `json:"isMarked" comment:"星标（1-已标记，0-未标记）"`
//	ObjectCid      string    `json:"objectCid" comment:"对象CID"`
//	CreatedAt      time.Time `json:"createdAt" comment:"创建时间"`
//	UpdatedAt      time.Time `json:"updatedAt" comment:"最后更新时间"`
//	CreatedDate    string    `json:"createdDate" comment:"创建日期"`
//}

// endregion CAR Import

//func makeBar(req *model.CarFileUploadReq) *pb.ProgressBar {
//	objectSize := int(req.ObjectSize)
//	bar := pb.New(objectSize).
//
//	bar := pb.New(int(sourceSize)).SetUnits(pb.U_BYTES).SetRefreshRate(time.Millisecond * 10)
//	bar.ShowSpeed = true
//	bar.
//	// show percents (by default already true)
//	bar.ShowPercent = true
//
//	// show bar (by default already true)
//	bar.ShowBar = true
//
//	bar.ShowCounters = true
//
//	bar.ShowTimeLeft = true
//}

// 生成CAR分片文件
func GenerateShardingCarFiles(req *model.CarFileUploadReq, shardingCarFileUploadReqs *[]model.CarFileUploadReq) error {
	fileDestination := req.FileDestination

	bigCarFile, err := os.Open(fileDestination)
	if err != nil {
		log.WithError(err).
			WithFields(logrus.Fields{
				"fileDestination": fileDestination,
			}).Error("Fail to open car file")

		return err
	}
	defer bigCarFile.Close()

	// CAR文件分片设置
	targetSize := sdkConfig.CarFileShardingThreshold //3145728 //1024 * 1024     // 1MiB chunks
	//strategy := carbites.Treewalk // also carbites.Treewalk
	//spltr, _ := carbites.Split(bigCarFile, targetSize, strategy)
	spltr, _ := carbites.NewTreewalkSplitterFromPath(fileDestination, targetSize)

	//shardingCarFileDestinationList := []string{}
	shardingNo := 1
	//var shardingNo int = 1

	for {
		car, err := spltr.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			//fmt.Printf("Error:%+v\n", err)
			log.WithError(err).
				WithFields(logrus.Fields{
					"shardingNo": shardingNo,
				}).Error("Fail to generate sharding car file")
			return err
		}

		//fmt.Printf("gernerate chunk shardingNo:%d start\n", shardingNo)
		bytes, err := io.ReadAll(car)
		if err != nil {
			//fmt.Printf("Error:%+v\n", err)
			log.WithError(err).
				WithFields(logrus.Fields{
					"shardingNo": shardingNo,
				}).Error("Fail to generate sharding car file")
			return err
		}

		// 设置文件名称
		filename := fmt.Sprintf("_chunk.c%d", shardingNo)
		//shardingFileDestination := generateFileName(utils.CurrentDate()+"_", filename)
		shardingFileDestination := strings.Replace(fileDestination, filepath.Ext(fileDestination), filename, 1)
		//shardingCarFileDestinationList = append(shardingCarFileDestinationList, shardingFileDestination)

		chunkSize := int64(len(bytes))

		// 生成分片文件
		//ioutil.WriteFile(fmt.Sprintf("chunk-%d.car", shardingNo), bytes, 0644)
		err = os.WriteFile(shardingFileDestination, bytes, 0644)
		if err != nil {
			//fmt.Printf("Error:%+v\n", err)
			log.WithError(err).
				WithFields(logrus.Fields{
					"shardingNo":              shardingNo,
					"shardingFileDestination": shardingFileDestination,
				}).Error("Fail to generate sharding car file")
			return err
		}
		//fmt.Printf("gernerate chunk shardingNo:%d end\n", shardingNo)

		//fmt.Printf("calculate chunk shardingSha256 start, shardingNo:%d\n", shardingNo)
		// 计算分片文件sha256
		shardingSha256, err := utils.GetFileSha256ByPath(shardingFileDestination)
		if err != nil {
			//fmt.Printf("Error:%+v\n", err)
			log.WithError(err).
				WithFields(logrus.Fields{
					"shardingNo":              shardingNo,
					"shardingFileDestination": shardingFileDestination,
				}).Error("Fail to calculate file sha256")
			return err
		}
		//carFileUploadReq.RawSha256 = shardingSha256
		//fmt.Printf("calculate chunk shardingSha256 end, shardingNo:%d\n", shardingNo)

		// 设置分片请求对象
		shardingCarFileUploadReq := model.CarFileUploadReq{}
		deepcopier.Copy(req).To(&shardingCarFileUploadReq)
		shardingCarFileUploadReq.FileDestination = shardingFileDestination
		shardingCarFileUploadReq.ShardingSha256 = shardingSha256
		shardingCarFileUploadReq.ShardingNo = shardingNo

		//// todo: remove it
		//rootLink := model.RootLink{}
		//parseCarDag(shardingFileDestination, &rootLink)
		//rootCid := rootLink.RootCid.String()
		//size := int64(rootLink.Size)
		//shardingCarFileUploadReq.CarFileCid = rootCid
		shardingCarFileUploadReq.ObjectSize = chunkSize

		*shardingCarFileUploadReqs = append(*shardingCarFileUploadReqs, shardingCarFileUploadReq)

		shardingNo++
	}

	// 分片失败
	shardingAmount := len(*shardingCarFileUploadReqs)
	if shardingAmount == 0 {
		//fmt.Printf("Error:%+v\n", err)
		log.WithError(err).
			WithFields(logrus.Fields{
				"shardingCarFileUploadReqs": shardingCarFileUploadReqs,
			}).Error("Fail to generate sharding car file")
		return sdkcode.ErrCarUploadFileChunkCarFileFail
	}

	req.ShardingAmount = shardingAmount

	return nil
}

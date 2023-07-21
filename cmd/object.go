package cmd

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/Code-Hex/pget"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-unixfsnode"
	"github.com/ipfs/go-unixfsnode/data"
	dagpb "github.com/ipld/go-codec-dagpb"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagcbor"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/node/basicnode"
	"github.com/ipld/go-ipld-prime/storage/memstore"
	"github.com/multiformats/go-multicodec"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	chainstoragesdk "github.com/solarfs/go-chainstorage-sdk"
	sdkcode "github.com/solarfs/go-chainstorage-sdk/code"
	"github.com/solarfs/go-chainstorage-sdk/consts"
	"github.com/solarfs/go-chainstorage-sdk/model"
	"github.com/solarfs/go-chainstorage-sdk/utils"
	"github.com/spf13/cobra"
	"github.com/ulule/deepcopier"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"text/template"
	"time"
)

func init() {
	//objectListCmd.Flags().StringP("Bucket", "b", "", "桶名称")
	//objectListCmd.Flags().StringP("Object", "r", "", "对象名称")
	//objectListCmd.Flags().StringP("Cid", "c", "", "Cid")
	//objectListCmd.Flags().IntP("Offset", "o", 10, "查询偏移量")
	//
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
	//
	//objectDownloadCmd.Flags().StringP("Bucket", "b", "", "桶名称")
	//objectDownloadCmd.Flags().StringP("Object", "o", "", "对象名称")
	//objectDownloadCmd.Flags().StringP("Cid", "c", "", "Cid")
	//objectDownloadCmd.Flags().BoolP("Target", "t", false, "输出路径")
}

// region Object List

//var objectListCmd = &cobra.Command{
//	Use:     "lso",
//	Short:   "lso",
//	Long:    "List object",
//	Example: "gcscmd ls cs://BUCKET [--name=<name>] [--cid=<cid>] [--Offset=<Offset>]",
//
//	Run: func(cmd *cobra.Command, args []string) {
//		//cmd.Help()
//		//fmt.Printf("%s %s\n", cmd.Name(), strconv.Itoa(offset))
//		objectListRun(cmd, args)
//	},
//}

func objectListRun(cmd *cobra.Command, args []string) {
	pageSize := 10
	pageIndex := 1

	// 桶名称
	bucketName := GetBucketName(args)
	if err := checkBucketName(bucketName); err != nil {
		Error(cmd, args, err)
	}

	// 对象CID
	objectCid, err := cmd.Flags().GetString("cid")
	if err != nil {
		Error(cmd, args, err)
	}

	if len(objectCid) != 0 {
		_, err = cid.Decode(objectCid)
		if err != nil {
			Error(cmd, args, err)
		}
	}

	// 对象名称
	objectName, err := cmd.Flags().GetString("name")
	if err != nil {
		Error(cmd, args, err)
	}

	if len(objectName) != 0 {
		if err := checkObjectName(objectName); err != nil {
			Error(cmd, args, err)
		}
	}

	// 设置参数
	// todo: Cid和name谁优先？
	objectItem := objectCid
	if len(objectName) != 0 {
		objectItem = objectName
	}

	// 查询偏移量
	//offset := viper.GetInt("cli.listOffset")
	offset := cliConfig.ListOffset
	if offset > 0 || offset <= 1000 {
		pageSize = offset
	}

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

	// 列出桶对象
	response, err := sdk.Object.GetObjectList(bucketId, objectItem, pageSize, pageIndex)
	if err != nil {
		Error(cmd, args, err)
	}

	objectListRunOutput(cmd, args, response)
}

func objectListRunOutput(cmd *cobra.Command, args []string, resp model.ObjectPageResponse) {
	code := int(resp.Code)
	if code != http.StatusOK {
		Error(cmd, args, errors.New(resp.Msg))

		//err := errors.Errorf("code:%d, message:&s\n", resp.Code, resp.Msg)
		//if code == sdkcode.ErrInvalidBucketId.Code() {
		//	err = errors.Errorf("bucket can't be found")
		//}

		//Error(cmd, args, err)
	}

	respData := resp.Data
	objectListOutput := ObjectListOutput{
		RequestId: resp.RequestId,
		Code:      resp.Code,
		Msg:       resp.Msg,
		Status:    resp.Status,
		Count:     respData.Count,
		PageIndex: respData.PageIndex,
		PageSize:  respData.PageSize,
		List:      []ObjectOutput{},
	}

	if len(respData.List) > 0 {
		for i := range respData.List {
			objectOutput := ObjectOutput{}
			deepcopier.Copy(respData.List[i]).To(&objectOutput)

			// 创建时间
			objectOutput.CreatedDate = objectOutput.CreatedAt.Format("2006-01-02")

			// 对象大小
			objectOutput.FormatObjectSize = convertSizeUnit(objectOutput.ObjectSize)

			objectListOutput.List = append(objectListOutput.List, objectOutput)
		}
	}

	templateContent := `
total {{.Count}}
{{- if eq (len .List) 0}}
Status: {{.Code}}
{{else}}
{{- range .List}}
{{.ObjectCid}} {{.FormatObjectSize}} {{.CreatedDate}} {{.ObjectName}}
{{- end}}
{{end}}`

	t, err := template.New("objectListTemplate").Parse(templateContent)
	if err != nil {
		Error(cmd, args, err)
	}

	err = t.Execute(os.Stdout, objectListOutput)
	if err != nil {
		Error(cmd, args, err)
	}
}

type ObjectListOutput struct {
	RequestId string         `json:"requestId,omitempty"`
	Code      int32          `json:"code,omitempty"`
	Msg       string         `json:"msg,omitempty"`
	Status    string         `json:"status,omitempty"`
	Count     int            `json:"count,omitempty"`
	PageIndex int            `json:"pageIndex,omitempty"`
	PageSize  int            `json:"pageSize,omitempty"`
	List      []ObjectOutput `json:"list,omitempty"`
}

type ObjectOutput struct {
	Id               int       `json:"id" comment:"对象ID"`
	BucketId         int       `json:"bucketId" comment:"桶主键"`
	ObjectName       string    `json:"objectName" comment:"对象名称（255字限制）"`
	ObjectTypeCode   int       `json:"objectTypeCode" comment:"对象类型编码"`
	ObjectSize       int64     `json:"objectSize" comment:"对象大小（字节）"`
	IsMarked         int       `json:"isMarked" comment:"星标（1-已标记，0-未标记）"`
	ObjectCid        string    `json:"objectCid" comment:"对象CID"`
	CreatedAt        time.Time `json:"createdAt" comment:"创建时间"`
	UpdatedAt        time.Time `json:"updatedAt" comment:"最后更新时间"`
	CreatedDate      string    `json:"createdDate" comment:"创建日期"`
	FormatObjectSize string    `json:"formatObjectSize" comment:"格式化对象大小"`
}

// endregion Object List

// region Object Rename

//var objectRenameCmd = &cobra.Command{
//	Use:     "rn",
//	Short:   "rn",
//	Long:    "rename object",
//	Example: "gcscmd rn cs://BUCKET] [--name=<name>] [--cid=<cid>] [--rename=<rename>] [--force]",
//
//	Run: func(cmd *cobra.Command, args []string) {
//		objectRenameRun(cmd, args)
//	},
//}

func objectRenameRun(cmd *cobra.Command, args []string) {
	// 桶名称
	bucketName := GetBucketName(args)
	if err := checkBucketName(bucketName); err != nil {
		Error(cmd, args, err)
	}

	// 对象CID
	objectCid, err := cmd.Flags().GetString("cid")
	if err != nil {
		Error(cmd, args, err)
	}

	if len(objectCid) != 0 {
		_, err = cid.Decode(objectCid)
		if err != nil {
			Error(cmd, args, err)
		}
	}

	// 对象名称
	objectName, err := cmd.Flags().GetString("name")
	if err != nil {
		Error(cmd, args, err)
	}

	if len(objectName) != 0 {
		if err := checkObjectName(objectName); err != nil {
			Error(cmd, args, err)
		}
	}

	if len(objectCid) == 0 && len(objectName) == 0 {
		Error(cmd, args, errors.New("please specify the name or cid"))
	}

	// 重命名
	rename, err := cmd.Flags().GetString("rename")
	if err != nil {
		Error(cmd, args, err)
	}

	if err := checkObjectName(rename); err != nil {
		Error(cmd, args, err)
	}

	//// todo: return succeed?
	//if rename == objectName {
	//	Error(cmd, args, errors.New("the new name of object can't be equal to the raw name of object"))
	//}

	// 强制覆盖
	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		Error(cmd, args, err)
	}

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

	// 确认对象数据有效性
	objectId := 0
	// todo: Cid和name谁优先？
	if len(objectCid) != 0 {
		pageSize := 1000
		pageIndex := 1
		respObject, err := sdk.Object.GetObjectList(bucketId, objectCid, pageSize, pageIndex)
		if err != nil {
			Error(cmd, args, err)
		}

		code = int(respObject.Code)
		if code != http.StatusOK {
			Error(cmd, args, errors.New(respObject.Msg))
		}

		count := respObject.Data.Count
		if count == 1 {
			objectData := respObject.Data.List[0]
			objectId = objectData.Id
			objectName = objectData.ObjectName
		} else if count == 0 {
			Error(cmd, args, sdkcode.ErrObjectNotFound)
		} else if count > 1 {
			// todo: please use name query?
			//Error(cmd, args, errors.New("Error: Multiple objects match this query, cannot perform this operation, please use cid query\n"))
			Error(cmd, args, errors.New("Error: Multiple objects match this query, cannot perform this operation, please use name query\n"))
		}
	} else {
		respObject, err := sdk.Object.GetObjectByName(bucketId, objectName)
		if err != nil {
			Error(cmd, args, err)
		}

		code = int(respObject.Code)
		if code != http.StatusOK {
			Error(cmd, args, errors.New(respObject.Msg))
		}

		// 对象ID
		objectId = respObject.Data.Id
	}

	response := model.ObjectRenameResponse{}
	// 重名名与对象名称相同，直接返回成功
	if rename == objectName {
		response.Code = http.StatusOK
		//Error(cmd, args, errors.New("the new name of object can't be equal to the raw name of object"))
	} else {
		// 重命名对象
		response, err = sdk.Object.RenameObject(objectId, rename, force)
		if err != nil {
			Error(cmd, args, err)
		}
	}

	objectRenameRunOutput(cmd, args, response)
}

func objectRenameRunOutput(cmd *cobra.Command, args []string, resp model.ObjectRenameResponse) {
	respCode := int(resp.Code)

	if respCode == sdkcode.ErrObjectNameConflictInBucket.Code() {
		err := errors.New("Error: conflicting rename filename, add --force to confirm overwrite\n")
		Error(cmd, args, err)
	} else if respCode != http.StatusOK {
		Error(cmd, args, errors.New(resp.Msg))
	}

	objectRenameOutput := ObjectRenameOutput{
		RequestId: resp.RequestId,
		Code:      resp.Code,
		Msg:       resp.Msg,
		Status:    resp.Status,
	}

	templateContent := `
Succeed
Status: {{.Code}}
`

	t, err := template.New("objectRenameTemplate").Parse(templateContent)
	if err != nil {
		Error(cmd, args, err)
	}

	err = t.Execute(os.Stdout, objectRenameOutput)
	if err != nil {
		Error(cmd, args, err)
	}
}

type ObjectRenameOutput struct {
	RequestId string       `json:"requestId,omitempty"`
	Code      int32        `json:"code,omitempty"`
	Msg       string       `json:"msg,omitempty"`
	Status    string       `json:"status,omitempty"`
	Data      ObjectOutput `json:"objectOutput,omitempty"`
}

// endregion Object Rename

// region Object Remove

//var objectRemoveCmd = &cobra.Command{
//	Use:     "rmo",
//	Short:   "rmo",
//	Long:    "remove object",
//	Example: "gcscmd rmo cs://BUCKET] [--name=<name>] [--cid=<cid>] [--remove=<remove>] [--force]",
//
//	Run: func(cmd *cobra.Command, args []string) {
//		objectRemoveRun(cmd, args)
//	},
//}

func objectRemoveRun(cmd *cobra.Command, args []string) {
	// 桶名称
	bucketName := GetBucketName(args)
	if err := checkBucketName(bucketName); err != nil {
		Error(cmd, args, err)
	}

	// 对象CID
	objectCid, err := cmd.Flags().GetString("cid")
	if err != nil {
		Error(cmd, args, err)
	}

	if len(objectCid) != 0 {
		_, err = cid.Decode(objectCid)
		if err != nil {
			Error(cmd, args, err)
		}
	}

	// 对象名称
	objectName, err := cmd.Flags().GetString("name")
	if err != nil {
		Error(cmd, args, err)
	}

	if len(objectName) != 0 {
		if err := checkObjectName(objectName); err != nil {
			Error(cmd, args, err)
		}
	}

	// 强制覆盖
	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		Error(cmd, args, err)
	}

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

	// todo: Cid和name谁优先？
	objectItem := objectCid
	if len(objectName) != 0 {
		objectItem = objectName
	}

	// 确认对象数据有效性
	var objectIdList []int
	pageSize := 1000
	pageIndex := 1
	respObjectList, err := sdk.Object.GetObjectList(bucketId, objectItem, pageSize, pageIndex)
	if err != nil {
		Error(cmd, args, err)
	}

	code = int(respObjectList.Code)
	if code != http.StatusOK {
		Error(cmd, args, errors.New(respObjectList.Msg))
	}

	count := respObjectList.Data.Count
	if count == 1 {
		if len(objectName) != 0 {
			// todo：模糊匹配还是精准匹配?
			rawObjectName := respObjectList.Data.List[0].ObjectName
			if !force && rawObjectName != objectName {
				Error(cmd, args, sdkcode.ErrObjectNotFound)
			}
		}

		objectId := respObjectList.Data.List[0].Id
		objectIdList = []int{objectId}
	} else if count == 0 {
		Error(cmd, args, sdkcode.ErrObjectNotFound)
	} else if count > 1 {
		if !force {
			Error(cmd, args, errors.New("Error: multiple object  are matching this query, add --force to confirm the bulk removal\n"))
		}

		for i := range respObjectList.Data.List {
			objectId := respObjectList.Data.List[i].Id
			objectIdList = append(objectIdList, objectId)
		}
	}

	// 重命名对象
	response, err := sdk.Object.RemoveObject(objectIdList)
	if err != nil {
		Error(cmd, args, err)
	}

	objectRemoveRunOutput(cmd, args, response)
}

func objectRemoveRunOutput(cmd *cobra.Command, args []string, resp model.ObjectRemoveResponse) {
	respCode := int(resp.Code)

	if respCode != http.StatusOK {
		Error(cmd, args, errors.New(resp.Msg))
	}

	objectRemoveOutput := ObjectRemoveOutput{
		RequestId: resp.RequestId,
		Code:      resp.Code,
		Msg:       resp.Msg,
		Status:    resp.Status,
	}

	//	删除对象
	//	通过命令删除固定桶内对象
	//
	//	模版
	//
	//	gcscmd rm cs://[BUCKET] [--name=<name>] [--cid=<cid>] [--force]
	//	BUCKET
	//
	//	桶名称
	//
	//	cid
	//
	//	添加对应的 CID
	//
	//	name
	//
	//	对象名
	//
	//	force
	//
	//	无添加筛选条件或命中多的对象时需要添加
	//
	//	命令行例子
	//
	//	清空桶
	//
	//	gcscmd rm cs://bbb --force
	//	使用对象名删除单文件
	//
	//	gcscmd rm  cs://bbb --name Tarkov.mp4
	//	使用模糊查询删除对象
	//
	//	gcscmd rm  cs://bbb --name .mp4 --force
	//	使用对象名删除单目录
	//
	//	gcscmd rm  cs://bbb --name aaa
	//	使用CID删除单对象
	//
	//	gcscmd rm  cs://bbb --cid QmWgnG7pPjG31w328hZyALQ2BgW5aQrZyKpT47jVpn8CNo
	//	使用 CID 删除多个对象(命中多个对象时加)
	//
	//	gcscmd rm  cs://bbb --cid QmWgnG7pPjG31w328hZyALQ2BgW5aQrZyKpT47jVpn8CNo --force
	//	响应
	//
	//	成功
	//
	//Status: 200
	//	多对象没有添加 force
	//
	//Error: multiple object  are matching this query, add --force to confirm the bulk removal

	templateContent := `
Succeed
Status: {{.Code}}
`

	t, err := template.New("objectRemoveTemplate").Parse(templateContent)
	if err != nil {
		Error(cmd, args, err)
	}

	err = t.Execute(os.Stdout, objectRemoveOutput)
	if err != nil {
		Error(cmd, args, err)
	}
}

type ObjectRemoveOutput struct {
	RequestId string       `json:"requestId,omitempty"`
	Code      int32        `json:"code,omitempty"`
	Msg       string       `json:"msg,omitempty"`
	Status    string       `json:"status,omitempty"`
	Data      ObjectOutput `json:"objectOutput,omitempty"`
}

// endregion Object Remove

// region Object Download

//var objectDownloadCmd = &cobra.Command{
//	Use:     "get",
//	Short:   "get",
//	Long:    "download object",
//	Example: "gcscmd get cs://BUCKET [--name=<name>] [--cid=<cid>]",
//
//	Run: func(cmd *cobra.Command, args []string) {
//		objectDownloadRun(cmd, args)
//	},
//}

func objectDownloadRun(cmd *cobra.Command, args []string) {
	// 桶名称
	bucketName := GetBucketName(args)
	if err := checkBucketName(bucketName); err != nil {
		Error(cmd, args, err)
	}

	// 对象CID
	objectCid, err := cmd.Flags().GetString("cid")
	if err != nil {
		Error(cmd, args, err)
	}

	if len(objectCid) != 0 {
		_, err = cid.Decode(objectCid)
		if err != nil {
			Error(cmd, args, err)
		}
	}

	// 对象名称
	objectName, err := cmd.Flags().GetString("name")
	if err != nil {
		Error(cmd, args, err)
	}

	if len(objectName) != 0 {
		if err := checkObjectName(objectName); err != nil {
			Error(cmd, args, err)
		}
	}

	if len(objectCid) == 0 && len(objectName) == 0 {
		Error(cmd, args, errors.New("please specify the name or cid"))
	}

	// 用户自定义目录
	downloadFolder, err := cmd.Flags().GetString("downloadFolder")
	if err != nil {
		Error(cmd, args, err)
	}

	// 用户自定义目录无效
	if len(downloadFolder) != 0 {
		fileInfo, err := os.Stat(downloadFolder)
		if err != nil {
			Error(cmd, args, errors.New("please specify the valid download folder"))
		}

		if !fileInfo.IsDir() {
			Error(cmd, args, errors.New("please specify the valid download folder"))
		}
	}

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

	// 确认对象数据有效性
	respObject := model.ObjectCreateResponse{}
	// todo: Cid和name谁优先？
	if len(objectCid) != 0 {
		pageSize := 1000
		pageIndex := 1
		respObjectList, err := sdk.Object.GetObjectList(bucketId, objectCid, pageSize, pageIndex)
		if err != nil {
			Error(cmd, args, err)
		}

		code = int(respObjectList.Code)
		if code != http.StatusOK {
			Error(cmd, args, errors.New(respObjectList.Msg))
		}

		count := respObjectList.Data.Count
		if count == 0 {
			Error(cmd, args, sdkcode.ErrObjectNotFound)
		} else if count > 1 {
			// todo: please use name query?
			//Error(cmd, args, errors.New("Error: Multiple objects match this query, cannot perform this operation, please use cid query\n"))
			Error(cmd, args, errors.New("Error: Multiple objects match this query, cannot perform this operation, please use name query\n"))
		}

		objectData := respObjectList.Data.List[0]
		objectName = objectData.ObjectName

		objectCreateResponse := model.ObjectCreateResponse{}
		deepcopier.Copy(&respObjectList).To(&objectCreateResponse)
		deepcopier.Copy(&objectData).To(&objectCreateResponse.Data)
		respObject = objectCreateResponse
	} else {
		respObject, err = sdk.Object.GetObjectByName(bucketId, objectName)
		if err != nil {
			Error(cmd, args, err)
		}

		code = int(respObject.Code)
		if code != http.StatusOK {
			Error(cmd, args, errors.New(respObject.Msg))
		}

		// 对象CID
		objectCid = respObject.Data.ObjectCid
	}

	isDir := respObject.Data.ObjectTypeCode == consts.ObjectTypeCodeDir
	if isDir {
		//downloadDirData(cmd, args, &respObject)
		downloadDirDataViaDagData(cmd, args, &respObject)
		return
	}

	//ipfsGateway := viper.GetString("cli.ipfsGateway")
	//downloadUrl := fmt.Sprintf("https://%s%s", ipfsGateway, objectCid)
	//downloadUrl := fmt.Sprintf(ipfsGateway, objectCid)
	//ipfsGateway := cliConfig.IpfsGateway
	//downloadUrl := ipfsGateway + objectCid

	err = downloadFile(objectCid, objectName, downloadFolder)
	if err != nil {
		Error(cmd, args, err)
	}

	objectDownloadRunOutput(cmd, args, respObject)
}

func downloadFile(objectCid string, objectName string, downloadFolder string) error {
	downloadUrl := generateDownloadUrl(objectCid, false)
	outputPath := objectName
	if len(downloadFolder) != 0 {
		outputPath = filepath.Join(downloadFolder, objectName)
	}

	cli := pget.New()
	cli.URLs = []string{downloadUrl}
	cli.Output = outputPath
	version := ""
	downloadArgs := []string{"-t", "10"}
	if err := cli.Run(context.Background(), version, downloadArgs); err != nil {
		//if cli.Trace {
		//	fmt.Fprintf(os.Stderr, "Error:\n%+v\n", err)
		//} else {
		//	fmt.Fprintf(os.Stderr, "Error:\n  %v\n", err)
		//}
		//Error(cmd, args, err)

		log.WithError(err).
			WithFields(logrus.Fields{
				"objectCid":      objectCid,
				"objectName":     objectName,
				"downloadFolder": downloadFolder,
				"outputPath":     outputPath,
			}).Error("Fail to download file")
		//fmt.Printf("Error:%+v\n", err)

		return err
	}

	log.WithFields(logrus.Fields{
		"objectCid":      objectCid,
		"objectName":     objectName,
		"downloadFolder": downloadFolder,
		"outputPath":     outputPath,
	}).Info("Success to download file")

	return nil
}

func objectDownloadRunOutput(cmd *cobra.Command, args []string, resp model.ObjectCreateResponse) {
	respCode := int(resp.Code)

	if respCode != http.StatusOK {
		Error(cmd, args, errors.New(resp.Msg))
	}

	objectDownloadOutput := ObjectDownloadOutput{
		RequestId: resp.RequestId,
		Code:      resp.Code,
		Msg:       resp.Msg,
		Status:    resp.Status,
	}

	err := deepcopier.Copy(&resp.Data).To(&objectDownloadOutput.Data)
	if err != nil {
		Error(cmd, args, err)
	}

	templateContent := `
CID: {{.Data.ObjectCid}}
Name: {{.Data.ObjectName}}
`

	t, err := template.New("objectDownloadTemplate").Parse(templateContent)
	if err != nil {
		Error(cmd, args, err)
	}

	err = t.Execute(os.Stdout, objectDownloadOutput)
	if err != nil {
		Error(cmd, args, err)
	}
}

type ObjectDownloadOutput struct {
	RequestId string       `json:"requestId,omitempty"`
	Code      int32        `json:"code,omitempty"`
	Msg       string       `json:"msg,omitempty"`
	Status    string       `json:"status,omitempty"`
	Data      ObjectOutput `json:"objectOutput,omitempty"`
}

func downloadDirDataFromCar(cmd *cobra.Command, args []string, respObject *model.ObjectCreateResponse) {
	objectCid := respObject.Data.ObjectCid
	objectName := respObject.Data.ObjectName
	downloadUrl := generateDownloadUrl(objectCid, true)

	sdk, err := chainstoragesdk.New(&appConfig)
	if err != nil {
		Error(cmd, args, err)
	}

	outputPath := sdk.Car.GenerateTempFileName(utils.CurrentDate()+"_", ".tmp")
	cli := pget.New()
	cli.URLs = []string{downloadUrl}
	cli.Output = outputPath
	version := ""
	downloadArgs := []string{"-t", "10"}
	if err := cli.Run(context.Background(), version, downloadArgs); err != nil {
		//if cli.Trace {
		//	fmt.Fprintf(os.Stderr, "Error:\n%+v\n", err)
		//} else {
		//	fmt.Fprintf(os.Stderr, "Error:\n  %v\n", err)
		//}
		Error(cmd, args, err)
	}

	// 用户自定义目录
	downloadDir, err := cmd.Flags().GetString("downloadfolder")
	if err != nil {
		Error(cmd, args, err)
	}

	// make data folder
	dirDestination := objectName
	if len(downloadDir) != 0 {
		dirDestination = filepath.Join(downloadDir, objectName)
	}

	// Check if the folder exists
	if _, err := os.Stat(dirDestination); os.IsNotExist(err) {
		// Folder does not exist, create a new folder
		err := os.MkdirAll(dirDestination, 0755)
		if err != nil {
			fmt.Println("Failed to create folder:", err)
			Error(cmd, args, err)
		}

		fmt.Println("Folder created:", dirDestination)
	}

	// extract car, delete temp car, override file
	err = sdk.Car.ExtractCarFile(outputPath, dirDestination)
	if err != nil {
		Error(cmd, args, err)
	}
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
	}(outputPath)

	objectDownloadRunOutput(cmd, args, *respObject)
}

func downloadDirData(cmd *cobra.Command, args []string, respObject *model.ObjectCreateResponse) {
	objectCid := respObject.Data.ObjectCid
	objectName := respObject.Data.ObjectName

	// 用户自定义目录
	downloadDir, err := cmd.Flags().GetString("downloadfolder")
	if err != nil {
		Error(cmd, args, err)
	}

	// make data folder
	outputDir := objectName
	if len(downloadDir) != 0 {
		outputDir = filepath.Join(downloadDir, objectName)
	}

	// Check if the folder exists
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		// Folder does not exist, create a new folder
		err := os.MkdirAll(outputDir, 0755)

		if err != nil {
			//fmt.Println("Failed to create folder:", err)
			log.Error(err)
			Error(cmd, args, err)
		}

		log.WithFields(logrus.Fields{
			"objectCid":   objectCid,
			"objectName":  objectName,
			"downloadDir": downloadDir,
			"outputDir":   outputDir,
		}).Info("Success to create folder")
		//fmt.Println("Folder created:", outputDir)
	}

	//// extract car, delete temp car, override file
	//err = sdk.Car.ExtractCarFile(outputPath, outputDir)
	//if err != nil {
	//	Error(cmd, args, err)
	//}
	//defer func(fileDestination string) {
	//	//if !viper.GetBool("cli.cleanTmpData") {
	//	if !cliConfig.CleanTmpData {
	//		return
	//	}
	//
	//	err := os.Remove(fileDestination)
	//	if err != nil {
	//		log.WithError(err).
	//			WithFields(logrus.Fields{
	//				"fileDestination": fileDestination,
	//			}).Error("Fail to remove car file")
	//		//fmt.Printf("Error:%+v\n", err)
	//	}
	//}(outputPath)

	// todo: progress bar should be displayed based on total size? and need stat total amount of downloading

	err = extractRoot(objectCid, outputDir)
	if err != nil {
		log.WithError(err).
			WithFields(logrus.Fields{
				"objectCid": objectCid,
				"outputDir": outputDir,
			}).Error("Fail to extract data from folder")
		//fmt.Printf("Error:%+v\n", err)
		Error(cmd, args, err)
	}

	// get dag data from ipfs gateway, and extract root dir info

	// iterate dag data

	// get http header info and check if it could be downloaded

	// the senario could be downloaded, and download it directly to specify folder

	// otherwise, continue to get dag data from ipfs gateway, and extract root dir info, recursively

	objectDownloadRunOutput(cmd, args, *respObject)
}

func downloadDirDataViaDagData(cmd *cobra.Command, args []string, respObject *model.ObjectCreateResponse) {
	objectCid := respObject.Data.ObjectCid
	objectName := respObject.Data.ObjectName

	// 用户自定义目录
	downloadDir, err := cmd.Flags().GetString("downloadFolder")
	if err != nil {
		Error(cmd, args, err)
	}

	// make data folder
	outputDir := objectName
	if len(downloadDir) != 0 {
		outputDir = filepath.Join(downloadDir, objectName)
	}

	// Check if the folder exists
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		// Folder does not exist, create a new folder
		err := os.MkdirAll(outputDir, 0755)

		if err != nil {
			//fmt.Println("Failed to create folder:", err)
			log.Error(err)
			Error(cmd, args, err)
		}

		log.WithFields(logrus.Fields{
			"objectCid":   objectCid,
			"objectName":  objectName,
			"downloadDir": downloadDir,
			"outputDir":   outputDir,
		}).Info("Success to create folder")
		//fmt.Println("Folder created:", outputDir)
	}

	nodes, err := extractFileNodes(objectCid, objectName, outputDir)
	if err != nil {
		log.WithError(err).
			WithFields(logrus.Fields{
				"objectCid":  objectCid,
				"objectName": objectName,
				"outputDir":  outputDir,
			}).Error("Fail to extract the info of file nodes from dag")
		//fmt.Printf("Error:%+v\n", err)
		Error(cmd, args, err)

		return
	}

	failAmount := 0
	totalAmount := 0
	failSize := int64(0)
	totalSize := int64(0)

	fmt.Println()
	fmt.Println("_/_/_/_/_/_/_/_/_/_/ download folder data start _/_/_/_/_/_/_/_/_/_/")
	fmt.Println("cid:", objectCid)
	fmt.Println("name:", objectName)
	for _, v := range nodes {
		cid := v.cid
		name := v.name
		relativePath := v.relativePath
		log.WithFields(logrus.Fields{
			"objectCid":    objectCid,
			"cid":          cid,
			"name":         name,
			"outputDir":    outputDir,
			"relativePath": relativePath,
		}).Info("download file in folder")

		// Check if the folder exists
		if _, err := os.Stat(relativePath); os.IsNotExist(err) {
			// Folder does not exist, create a new folder
			err := os.MkdirAll(relativePath, 0755)

			if err != nil {
				//fmt.Println("Failed to create folder:", err)
				log.Error(err)
				Error(cmd, args, err)
			}

			log.WithFields(logrus.Fields{
				"objectCid":    objectCid,
				"cid":          cid,
				"name":         name,
				"outputDir":    outputDir,
				"relativePath": relativePath,
			}).Info("Success to create folder")
			//fmt.Println("Folder created:", outputDir)
		}

		if err := downloadFile(cid, name, relativePath); err != nil {
			log.WithError(err).
				WithFields(logrus.Fields{
					"objectCid":    objectCid,
					"objectName":   objectName,
					"cid":          cid,
					"name":         name,
					"outputDir":    outputDir,
					"relativePath": relativePath,
				}).Error("Fail to download file in folder")
			//fmt.Printf("Error:%+v\n", err)
			Error(cmd, args, err)
			failSize += v.size
			failAmount++
			//return
		}

		totalSize += v.size
		totalAmount++
	}

	fmt.Println("download total amount:", totalAmount, ",total fail amount:", failAmount)
	fmt.Println("download total size:", totalSize, ",total fail size:", failSize)
	fmt.Println("_/_/_/_/_/_/_/_/_/_/ download folder data end _/_/_/_/_/_/_/_/_/_/")
	fmt.Println()

	//// extract car, delete temp car, override file
	//err = sdk.Car.ExtractCarFile(outputPath, outputDir)
	//if err != nil {
	//	Error(cmd, args, err)
	//}
	//defer func(fileDestination string) {
	//	//if !viper.GetBool("cli.cleanTmpData") {
	//	if !cliConfig.CleanTmpData {
	//		return
	//	}
	//
	//	err := os.Remove(fileDestination)
	//	if err != nil {
	//		log.WithError(err).
	//			WithFields(logrus.Fields{
	//				"fileDestination": fileDestination,
	//			}).Error("Fail to remove car file")
	//		//fmt.Printf("Error:%+v\n", err)
	//	}
	//}(outputPath)

	// todo: progress bar should be displayed based on total size? and need stat total amount of downloading

	//err = extractRoot(objectCid, outputDir)
	//if err != nil {
	//	log.WithError(err).
	//		WithFields(logrus.Fields{
	//			"objectCid": objectCid,
	//			"outputDir": outputDir,
	//		}).Error("Fail to extract data from folder")
	//	//fmt.Printf("Error:%+v\n", err)
	//	Error(cmd, args, err)
	//}

	// get dag data from ipfs gateway, and extract root dir info

	// iterate dag data

	// get http header info and check if it could be downloaded

	// the senario could be downloaded, and download it directly to specify folder

	// otherwise, continue to get dag data from ipfs gateway, and extract root dir info, recursively

	objectDownloadRunOutput(cmd, args, *respObject)
}

// generate downloading url
func generateDownloadUrl(objectCid string, getDag bool) string {
	downloadUrl := ""
	ipfsGateway := cliConfig.IpfsGateway

	base, err := url.Parse(ipfsGateway)
	if err != nil {
		log.Error(err)
	}

	ref, err := url.Parse(objectCid)
	if err != nil {
		log.Error(err)
	}

	if getDag {
		queryString := ref.Query()
		//queryString.Set("format", "dag-cbor")
		queryString.Set("format", "dag-json")
		ref.RawQuery = queryString.Encode()
	}

	u := base.ResolveReference(ref)
	downloadUrl = u.String()
	//fmt.Println(downloadUrl)

	return downloadUrl
}

// parse the dag data to IPLD.Node object
func parseDagData(objectCid string) (ipld.Node, ipld.Node, error) {
	url := generateDownloadUrl(objectCid, true)
	if len(url) == 0 {
		return nil, nil, fmt.Errorf("fail to generate download url")
	}

	dagData, err := getDagData(url)
	if err != nil {
		//fmt.Println("Error:", err)
		log.Error(err)
		return nil, nil, err
	}

	reader := bytes.NewReader(dagData)
	opts := dagcbor.DecodeOptions{
		AllowLinks: true,
		//ExperimentalDeterminism: true,
		DontParseBeyondEnd: true,
	}

	//builder := basicnode.Prototype.Map.NewBuilder()
	builder := basicnode.Prototype.Any.NewBuilder()
	err = opts.Decode(builder, reader)
	if err != nil {
		//fmt.Println("Failed to decode dag-cbor data:", err)
		log.Error(err)
		return nil, nil, err
	}
	node := builder.Build()

	//if node.Kind() == ipld.Kind_Bytes {
	if node.Kind() != ipld.Kind_Map {
		return nil, nil, ErrNotDir
	}

	linkSystem := cidlink.DefaultLinkSystem()
	storage := &memstore.Store{}
	linkSystem.TrustedStorage = true
	linkSystem.SetReadStorage(storage)
	linkSystem.SetWriteStorage(storage)

	// Store the IPLD node and get link back.
	gotLink, err := linkSystem.Store(ipld.LinkContext{}, cidlink.LinkPrototype{
		Prefix: cid.Prefix{
			Version:  1,
			Codec:    uint64(multicodec.DagPb),
			MhType:   uint64(multicodec.Sha2_256),
			MhLength: -1,
		},
	}, node)
	if err != nil {
		//fmt.Println("Failed to store dag data to Link System:", err)
		log.Error(err)
		return nil, nil, err
	}
	gotCidlink := gotLink.(cidlink.Link)

	pbn, err := linkSystem.Load(ipld.LinkContext{}, gotCidlink, dagpb.Type.PBNode)
	if err != nil {
		//fmt.Println("Failed to load dag data from Link System:", err)
		log.Error(err)
		return nil, nil, err
	}
	pbnode := pbn.(dagpb.PBNode)

	ufn, err := unixfsnode.Reify(ipld.LinkContext{}, pbnode, &linkSystem)
	if err != nil {
		//fmt.Println("Failed to reify unixfsnode:", err)
		log.Error(err)
		return nil, nil, err
	}

	return pbnode, ufn, nil
}

// extract the root of folder
func extractRoot(objectCid, outputDir string) error {
	//outputDir = "myfolder"
	//objectCid = "bafybeibhrzxon75wfciqa3jcdbrfdepvbbjpn67wffxrl47i74jrkl2ewi"
	pbnode, ufn, err := parseDagData(objectCid)
	if err != nil {
		log.Error(err)
		return err
	}

	// check if it is folder
	if ufn.Kind() != ipld.Kind_Map {
		return ErrNotDir
	}

	links, err := pbnode.LookupByString("Links")
	if err != nil {
		//fmt.Println("Error:", err)
		log.Error(err)
		return err
	}

	li := links.ListIterator()
	for !li.Done() {
		_, v, err := li.Next()
		if err != nil {
			//fmt.Println("Error:", err)
			log.Error(err)
			return err
		}

		pbLink := v.(dagpb.PBLink)
		name, err := pbLink.Name.AsNode().AsString()
		if err != nil {
			log.Error(err)
			return err
		}

		vl, err := pbLink.Hash.AsLink()
		if err != nil {
			log.Error(err)
			return err
		}

		//size, err := pbLink.Tsize.AsNode().AsInt()
		//if err != nil {
		//	log.Error(err)
		//	return err
		//}

		//fmt.Println("Name:", name)
		//fmt.Println("Hash:", vl.String())
		//fmt.Println("Tsize:", size)
		//fmt.Println(pbLink)

		cid := vl.String()
		//childNode, err := extractFolder(cid, outputDir, name)
		_, err = extractFolder(cid, outputDir, name)
		if err != nil {
			if !errors.Is(err, ErrNotDir) {
				log.Error(err)
				return fmt.Errorf("%s: %w", cid, err)
			}

			//// if it's not a directory, it's a file.
			//ufsData, err := childNode.LookupByString("Data")
			//if err != nil {
			//	return err
			//}
			//
			//ufsBytes, err := ufsData.AsBytes()
			//if err != nil {
			//	return err
			//}
			//
			//ufsNode, err := data.DecodeUnixFSData(ufsBytes)
			//if err != nil {
			//	return err
			//}
			//
			//if ufsNode.DataType.Int() == data.Data_File || ufsNode.DataType.Int() == data.Data_Raw {
			//	//fmt.Printf("download file(%s) to %s\n", name, outputDir)

			log.WithFields(logrus.Fields{
				"objectCid": objectCid,
				"cid":       cid,
				"name":      name,
				"outputDir": outputDir,
			}).Info("download file in folder")

			if err := downloadFile(cid, name, outputDir); err != nil {
				return err
			}
			//}
		}
	}

	return nil
}

// extract folder data
func extractFolder(objectCid, outputDir, objectName string) (ipld.Node, error) {
	//println("Cid:", objectCid)
	pbnode, ufn, err := parseDagData(objectCid)
	if err != nil {
		//log.Error(err)
		return nil, err
	}

	if ufn.Kind() != ipld.Kind_Map {
		return pbnode, ErrNotDir
	}

	downloadDir := outputDir
	outputDir = path.Join(outputDir, objectName)
	//fmt.Println("Make dir:", outputDir)
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		// Folder does not exist, create a new folder
		err := os.MkdirAll(outputDir, 0755)
		if err != nil {
			//fmt.Println("Failed to create folder:", err)
			log.Error(err)
			return pbnode, err
		}

		log.WithFields(logrus.Fields{
			"objectCid":   objectCid,
			"objectName":  objectName,
			"downloadDir": downloadDir,
			"outputDir":   outputDir,
		}).Info("Success to create folder")
		//fmt.Println("Folder created:", dirDestination)
	}

	links, err := pbnode.LookupByString("Links")
	if err != nil {
		//fmt.Println("Error:", err)
		log.Error(err)
		return pbnode, err
	}

	li := links.ListIterator()
	for !li.Done() {
		_, v, err := li.Next()
		if err != nil {
			//fmt.Println("Error:", err)
			log.Error(err)
			return pbnode, err
		}

		pbLink := v.(dagpb.PBLink)
		name, err := pbLink.Name.AsNode().AsString()
		if err != nil {
			log.Error(err)
			return pbnode, err
		}

		vl, err := pbLink.Hash.AsLink()
		if err != nil {
			log.Error(err)
			return pbnode, err
		}

		//size, err := pbLink.Tsize.AsNode().AsInt()
		//if err != nil {
		//	log.Error(err)
		//	return pbnode, err
		//}

		//fmt.Println("Name:", name)
		//fmt.Println("Hash:", vl.String())
		//fmt.Println("Tsize:", size)
		//fmt.Println(pbLink)

		cid := vl.String()
		//childNode, err := extractFolder(cid, outputDir, name)
		_, err = extractFolder(cid, outputDir, name)
		if err != nil {
			if !errors.Is(err, ErrNotDir) {
				log.Error(err)
				return pbnode, fmt.Errorf("%s: %w", cid, err)
			}

			//// if it's not a directory, it's a file.
			//ufsData, err := childNode.LookupByString("Data")
			//if err != nil {
			//	return pbnode, err
			//}
			//
			//ufsBytes, err := ufsData.AsBytes()
			//if err != nil {
			//	return pbnode, err
			//}
			//
			//ufsNode, err := data.DecodeUnixFSData(ufsBytes)
			//if err != nil {
			//	return pbnode, err
			//}
			//
			//if ufsNode.DataType.Int() == data.Data_File || ufsNode.DataType.Int() == data.Data_Raw {
			//	//fmt.Printf("download file(%s) to %s\n", name, outputDir)

			log.WithFields(logrus.Fields{
				"objectCid": objectCid,
				"cid":       cid,
				"name":      name,
				"outputDir": outputDir,
			}).Info("download file in folder")

			if err := downloadFile(cid, name, outputDir); err != nil {
				return pbnode, err
			}
			//}
		}
	}

	return pbnode, nil
}

// extract the info of file node from dags
func extractFileNodes(objectCid, objectName, outputDir string) ([]DagFileNode, error) {
	url := generateDownloadUrl(objectCid, true)

	dagData, err := getDagData(url)
	if err != nil {
		//fmt.Println("Error:", err)
		log.Error(err)
		return nil, err
	}

	reader := bytes.NewReader(dagData)

	bufioReader := bufio.NewReader(reader)
	opts := dagjson.DecodeOptions{
		ParseLinks:         true,
		ParseBytes:         true,
		DontParseBeyondEnd: true,
	}

	nb1 := basicnode.Prototype.Any.NewBuilder()
	err = opts.Decode(nb1, bufioReader)
	if err != nil {
		//fmt.Println("Failed to decode:", err)
		log.Error(err)
		return nil, err
	}

	node := nb1.Build()

	//fmt.Println("Kind():", node.Kind())
	// not found root dir, should considerate into single file?
	if node.Kind() == ipld.Kind_Bytes {
		//fmt.Println("Error:", ErrNotDir)
		log.Error(ErrNotDir)
		return nil, ErrNotDir
	}

	if node.Kind() != ipld.Kind_Map {
		//fmt.Println("Error:", ErrNotDir)
		log.Error(ErrNotDir)
		return nil, ErrNotDir
	}

	// interpret dagpb 'data' as unixfs data and look at type.
	ufsData, err := node.LookupByString("Data")
	if err != nil {
		//fmt.Println("Error:", err)
		log.Error(err)
		return nil, err
	}

	ufsBytes, err := ufsData.AsBytes()
	if err != nil {
		//fmt.Println("Error:", err)
		log.Error(err)
		return nil, err
	}

	ufsNode, err := data.DecodeUnixFSData(ufsBytes)
	if err != nil {
		//fmt.Println("Error:", err)
		log.Error(err)
		return nil, err
	}

	// check if it is a folder
	isDir := ufsNode.DataType.Int() == data.Data_Directory
	if !isDir {
		//fmt.Println("Error:", ErrNotDir)
		log.Error(ErrNotDir)
		return nil, ErrNotDir
	}

	var nodes []DagFileNode
	if node.Kind() == ipld.Kind_Map {
		links, err := node.LookupByString("Links")
		if err != nil {
			//fmt.Println("Error:", err)
			log.Error(err)
			return nil, err
		}

		li := links.ListIterator()
		for !li.Done() {
			_, v, err := li.Next()
			if err != nil {
				//fmt.Println("Error:", err)
				log.Error(err)
				return nil, err
			}

			if v.Kind() != ipld.Kind_Map {
				continue
			}

			hashNode, err := v.LookupByString("Hash")
			if err != nil {
				//fmt.Println("Error:", err)
				log.Error(err)
				return nil, err
			}

			cid, err := hashNode.AsLink()
			if err != nil {
				//fmt.Println("Error:", err)
				log.Error(err)
				return nil, err
			}

			nameNode, err := v.LookupByString("Name")
			if err != nil {
				//fmt.Println("Error:", err)
				log.Error(err)
				return nil, err
			}

			name, err := nameNode.AsString()
			if err != nil {
				//fmt.Println("Error:", err)
				log.Error(err)
				return nil, err
			}

			//sizeNode, err := v.LookupByString("Tsize")
			//if err != nil {
			//	//fmt.Println("Error:", err)
			//	log.Error(err)
			//	return nil, err
			//}
			//
			//size, err := sizeNode.AsInt()
			//if err != nil {
			//	//fmt.Println("Error:", err)
			//	log.Error(err)
			//	return nil, err
			//}

			//fmt.Println("Name:", name)
			//fmt.Println("Hash:", cid.String())
			//fmt.Println("Tsize:", size)
			//fmt.Println("Kind():", v.Kind())

			err = traverseDag(cid.String(), name, outputDir, &nodes)
			if err != nil {
				//fmt.Println("Error:", err)
				log.Error(err)
				return nil, err
			}
		}
	}

	return nodes, nil
}

// traverse dag
func traverseDag(objectCid, objectName, outputDir string, nodes *[]DagFileNode) error {
	url := generateDownloadUrl(objectCid, true)

	dagData, err := getDagData(url)
	if err != nil {
		//fmt.Println("Error:", err)
		log.Error(err)
		return err
	}

	reader := bytes.NewReader(dagData)
	opts := dagjson.DecodeOptions{
		ParseLinks:         true,
		ParseBytes:         true,
		DontParseBeyondEnd: true,
	}

	nb1 := basicnode.Prototype.Any.NewBuilder()
	err = opts.Decode(nb1, reader)
	if err != nil {
		//fmt.Println("Failed to decode:", err)
		log.Error(err)
		return err
	}

	node := nb1.Build()

	// it is a file
	if node.Kind() == ipld.Kind_Bytes {
		fileNode := DagFileNode{}
		fileNode.cid = objectCid
		fileNode.name = objectName

		bytes, err := node.AsBytes()
		if err != nil {
			//fmt.Println("Error:", err)
			log.Error(err)
			return err
		}
		fileNode.size = int64(len(bytes))

		fileNode.relativePath = outputDir
		*nodes = append(*nodes, fileNode)

		return nil
	}

	// interpret dagpb 'data' as unixfs data and look at type.
	ufsData, err := node.LookupByString("Data")
	if err != nil {
		//fmt.Println("Error:", err)
		log.Error(err)
		return err
	}

	ufsBytes, err := ufsData.AsBytes()
	if err != nil {
		//fmt.Println("Error:", err)
		log.Error(err)
		return err
	}

	ufsNode, err := data.DecodeUnixFSData(ufsBytes)
	if err != nil {
		//fmt.Println("Error:", err)
		log.Error(err)
		return err
	}

	switch ufsNode.DataType.Int() {
	case data.Data_Directory, data.Data_HAMTShard:
		outputDir = path.Join(outputDir, objectName)

	case data.Data_File, data.Data_Raw:
		fileNode := DagFileNode{}
		fileNode.cid = objectCid
		fileNode.name = objectName

		size, err := ufsNode.FileSize.AsNode().AsInt()
		if err != nil {
			//fmt.Println("Error:", err)
			log.Error(err)
			return err
		}

		fileNode.size = size
		fileNode.relativePath = outputDir
		*nodes = append(*nodes, fileNode)
		return nil

	case data.Data_Symlink:
		return nil

	default:
		err := fmt.Errorf("unknown unixfs type: %d", ufsNode.DataType.Int())
		//fmt.Println("Error:", err)
		log.Error(err)
		return err
	}

	if node.Kind() == ipld.Kind_Map {
		links, err := node.LookupByString("Links")
		if err != nil {
			//fmt.Println("Error:", err)
			log.Error(err)
			return err
		}

		li := links.ListIterator()
		for !li.Done() {
			_, v, err := li.Next()
			if err != nil {
				//fmt.Println("Error:", err)
				log.Error(err)
				return err
			}

			if v.Kind() != ipld.Kind_Map {
				continue
			}

			hashNode, err := v.LookupByString("Hash")
			if err != nil {
				//fmt.Println("Error:", err)
				log.Error(err)
				return err
			}

			cid, err := hashNode.AsLink()
			if err != nil {
				//fmt.Println("Error:", err)
				log.Error(err)
				return err
			}

			nameNode, err := v.LookupByString("Name")
			if err != nil {
				//fmt.Println("Error:", err)
				log.Error(err)
				return err
			}

			name, err := nameNode.AsString()
			if err != nil {
				//fmt.Println("Error:", err)
				log.Error(err)
				return err
			}

			//sizeNode, err := v.LookupByString("Tsize")
			//if err != nil {
			//	//fmt.Println("Error:", err)
			//	log.Error(err)
			//	return err
			//}
			//
			//size, err := sizeNode.AsInt()
			//if err != nil {
			//	//fmt.Println("Error:", err)
			//	log.Error(err)
			//	return err
			//}

			//fmt.Println("Name:", name)
			//fmt.Println("Hash:", cid.String())
			//fmt.Println("Tsize:", size)
			//fmt.Println("Kind():", v.Kind())

			err = traverseDag(cid.String(), name, outputDir, nodes)
			if err != nil {
				//fmt.Println("Error:", err)
				log.Error(err)
				return err
			}
		}
	}

	return nil
}

var ErrNotDir = fmt.Errorf("not a directory")

type DagFileNode struct {
	cid          string
	name         string
	size         int64
	relativePath string
}

// endregion Object Download

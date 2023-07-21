package cmd

import (
	"github.com/pkg/errors"
	chainstoragesdk "github.com/solarfs/go-chainstorage-sdk"
	sdkcode "github.com/solarfs/go-chainstorage-sdk/code"
	"github.com/solarfs/go-chainstorage-sdk/consts"
	"github.com/solarfs/go-chainstorage-sdk/model"
	"github.com/spf13/cobra"
	"github.com/ulule/deepcopier"
	"net/http"
	"os"
	"text/template"
	"time"
)

//
//var (
//	offset int
//)

func init() {
	//bucketListCmd.Flags().StringP("Bucket", "b", "", "桶名称")
	//bucketListCmd.Flags().IntP("Offset", "o", 10, "查询偏移量")
	//
	//bucketCreateCmd.Flags().StringP("Bucket", "b", "", "桶名称")
	//bucketCreateCmd.Flags().IntP("Storage", "s", 10001, "存储网络编码")
	//bucketCreateCmd.Flags().IntP("Principle", "p", 10001, "桶策略编码")
	//
	//bucketRemoveCmd.Flags().StringP("Bucket", "b", "", "桶名称")
	//bucketRemoveCmd.Flags().BoolP("Force", "f", false, "如果有数据，先清空再删除桶")
	//
	//bucketEmptyCmd.Flags().StringP("Bucket", "b", "", "桶名称")
}

// region Bucket List

//var bucketListCmd = &cobra.Command{
//	Use:     "ls",
//	Short:   "ls",
//	Long:    "List links from object or bucket",
//	Example: "gcscmd ls [--Offset=<Offset>]",
//
//	Run: func(cmd *cobra.Command, args []string) {
//		//cmd.Help()
//		//fmt.Printf("%s %s\n", cmd.Name(), strconv.Itoa(offset))
//		bucketListRun(cmd, args)
//	},
//}

func bucketListRun(cmd *cobra.Command, args []string) {
	bucketName := ""
	pageSize := 10
	pageIndex := 1

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

	// 列出桶对象
	response, err := sdk.Bucket.GetBucketList(bucketName, pageSize, pageIndex)
	if err != nil {
		Error(cmd, args, err)
	}

	bucketListRunOutput(cmd, args, response)
}

func bucketListRunOutput(cmd *cobra.Command, args []string, resp model.BucketPageResponse) {
	code := resp.Code
	if code != http.StatusOK {
		err := errors.Errorf("code:%d, message:%s\n", resp.Code, resp.Msg)
		Error(cmd, args, err)
	}

	respData := resp.Data
	bucketListOutput := BucketListOutput{
		RequestId: resp.RequestId,
		Code:      resp.Code,
		Msg:       resp.Msg,
		Status:    resp.Status,
		Count:     respData.Count,
		PageIndex: respData.PageIndex,
		PageSize:  respData.PageSize,
		List:      []BucketOutput{},
	}

	if len(respData.List) > 0 {
		for i := range respData.List {
			bucketOutput := BucketOutput{}
			deepcopier.Copy(respData.List[i]).To(&bucketOutput)

			// 存储网络
			bucketOutput.StorageNetwork = consts.StorageNetworkCodeMapping[bucketOutput.StorageNetworkCode]

			//// 桶策略
			//bucketOutput.BucketPrinciple = consts.BucketPrincipleCodeMapping[bucketOutput.BucketPrincipleCode]

			// 桶策略（英文）
			bucketOutput.BucketPrinciple = consts.BucketPrincipleCodeMappingEn[bucketOutput.BucketPrincipleCode]

			// 创建时间
			bucketOutput.CreatedDate = bucketOutput.CreatedAt.Format("2006-01-02")

			// 已使用空间
			bucketOutput.FormatUsedSpace = convertSizeUnit(bucketOutput.UsedSpace)

			bucketListOutput.List = append(bucketListOutput.List, bucketOutput)
		}
	}

	templateContent := `
total {{.Count}}
{{- if eq (len .List) 0}}
Status: {{.Code}}
{{- else}}
{{- range .List}}
{{.StorageNetwork}} {{.BucketPrinciple}} {{.FormatUsedSpace}} {{.ObjectAmount}} {{.CreatedDate}} {{.BucketName}}
{{- end}}
{{- end}}
`

	t, err := template.New("bucketListTemplate").Parse(templateContent)
	if err != nil {
		Error(cmd, args, err)
	}

	err = t.Execute(os.Stdout, bucketListOutput)
	if err != nil {
		Error(cmd, args, err)
	}
}

type BucketListOutput struct {
	RequestId string         `json:"requestId,omitempty"`
	Code      int32          `json:"code,omitempty"`
	Msg       string         `json:"msg,omitempty"`
	Status    string         `json:"status,omitempty"`
	Count     int            `json:"count,omitempty"`
	PageIndex int            `json:"pageIndex,omitempty"`
	PageSize  int            `json:"pageSize,omitempty"`
	List      []BucketOutput `json:"list,omitempty"`
}

type BucketOutput struct {
	Id                  int       `json:"id" comment:"桶ID"`
	BucketName          string    `json:"bucketName" comment:"桶名称（3-63字长度限制）"`
	StorageNetworkCode  int       `json:"storageNetworkCode" comment:"存储网络编码（10001-IPFS）"`
	BucketPrincipleCode int       `json:"bucketPrincipleCode" comment:"桶策略编码（10001-公开，10000-私有）"`
	UsedSpace           int64     `json:"usedSpace" comment:"已使用空间（字节）"`
	ObjectAmount        int       `json:"objectAmount" comment:"对象数量"`
	CreatedAt           time.Time `json:"createdAt" comment:"创建时间"`
	StorageNetwork      string    `json:"storageNetwork" comment:"存储网络（10001-IPFS）"`
	BucketPrinciple     string    `json:"bucketPrinciple" comment:"桶策略（10001-公开，10000-私有）"`
	CreatedDate         string    `json:"createdDate" comment:"创建日期"`
	FormatUsedSpace     string    `json:"formatUsedSpace" comment:"格式化已使用空间"`
}

// endregion Bucket List

// region Bucket Create

//var bucketCreateCmd = &cobra.Command{
//	Use:     "mb",
//	Short:   "mb",
//	Long:    "create bucket",
//	Example: "gcscmd mb cs://[BUCKET] [--storageNetworkCode=<storageNetworkCode>] [--bucketPrincipleCode=<bucketPrincipleCode>]",
//
//	Run: func(cmd *cobra.Command, args []string) {
//		//cmd.Help()
//		//fmt.Printf("%s %s\n", cmd.Name(), strconv.Itoa(offset))
//		bucketCreateRun(cmd, args)
//	},
//}

func bucketCreateRun(cmd *cobra.Command, args []string) {
	// 桶名称
	bucketName := GetBucketName(args)
	if err := checkBucketName(bucketName); err != nil {
		Error(cmd, args, err)
	}

	//bucketCreateCmd.Flags().IntP("Storage", "s", 10001, "存储网络编码")
	//bucketCreateCmd.Flags().IntP("Principle", "p", 10001, "桶策略编码")
	// 存储网络编码
	storageNetworkCode, err := cmd.Flags().GetInt("storageNetworkCode")
	if err != nil {
		Error(cmd, args, err)
	}

	if storageNetworkCode > 0 {
		_, exist := consts.StorageNetworkCodeMapping[storageNetworkCode]
		if !exist {
			err := errors.Errorf("invalid storage network code, %d", storageNetworkCode)
			Error(cmd, args, err)
		}
	}

	// 桶策略编码
	bucketPrincipleCode, err := cmd.Flags().GetInt("bucketPrincipleCode")
	if err != nil {
		Error(cmd, args, err)
	}

	if bucketPrincipleCode > 0 {
		_, exist := consts.BucketPrincipleCodeMapping[bucketPrincipleCode]
		if !exist {
			err := errors.Errorf("invalid bucket principle code, %d", bucketPrincipleCode)
			Error(cmd, args, err)
		}
	}

	sdk, err := chainstoragesdk.New(&appConfig)
	if err != nil {
		Error(cmd, args, err)
	}

	// 创建桶
	response, err := sdk.Bucket.CreateBucket(bucketName, storageNetworkCode, bucketPrincipleCode)
	if err != nil {
		Error(cmd, args, err)
	}

	bucketCreateRunOutput(cmd, args, response)
}

func bucketCreateRunOutput(cmd *cobra.Command, args []string, resp model.BucketCreateResponse) {
	code := resp.Code
	if code != http.StatusOK {
		Error(cmd, args, errors.New(resp.Msg))
	}

	respData := resp.Data
	bucketCreateOutput := BucketCreateOutput{
		RequestId: resp.RequestId,
		Code:      resp.Code,
		Msg:       resp.Msg,
		Status:    resp.Status,
	}

	bucketOutput := BucketOutput{}
	deepcopier.Copy(respData).To(&bucketOutput)

	// 存储网络
	bucketOutput.StorageNetwork = consts.StorageNetworkCodeMapping[bucketOutput.StorageNetworkCode]

	//// 桶策略
	//bucketOutput.BucketPrinciple = consts.BucketPrincipleCodeMapping[bucketOutput.BucketPrincipleCode]

	// 桶策略（英文）
	bucketOutput.BucketPrinciple = consts.BucketPrincipleCodeMappingEn[bucketOutput.BucketPrincipleCode]

	// 创建时间 todo: timezone
	//bucketOutput.CreatedDate = bucketOutput.CreatedAt.Format("2006-01-02")
	bucketOutput.CreatedDate = bucketOutput.CreatedAt.Format("2006-01-02T15:04:05-07:00")
	bucketCreateOutput.Data = bucketOutput
	//bucketOutput.CreatedAt.Format(time.RFC3339)

	templateContent := `
BUCKET: {{.BucketName}}
storageNetwork: {{.StorageNetwork}}
bucketPrinciple: {{.BucketPrinciple}}
createdAt: {{.CreatedDate}}
`

	t, err := template.New("bucketCreateTemplate").Parse(templateContent)
	if err != nil {
		Error(cmd, args, errors.New(resp.Msg))
	}

	err = t.Execute(os.Stdout, bucketOutput)
	if err != nil {
		Error(cmd, args, errors.New(resp.Msg))
	}
}

type BucketCreateOutput struct {
	RequestId string       `json:"requestId,omitempty"`
	Code      int32        `json:"code,omitempty"`
	Msg       string       `json:"msg,omitempty"`
	Status    string       `json:"status,omitempty"`
	Data      BucketOutput `json:"bucketOutput,omitempty"`
}

// endregion Bucket Create

// region Bucket Remove

//var bucketRemoveCmd = &cobra.Command{
//	Use:     "rb",
//	Short:   "rb",
//	Long:    "remove bucket",
//	Example: "gcscmd rb cs://[BUCKET] [--force]",
//
//	Run: func(cmd *cobra.Command, args []string) {
//		bucketRemoveRun(cmd, args)
//	},
//}

func bucketRemoveRun(cmd *cobra.Command, args []string) {
	// 桶名称
	bucketName := GetBucketName(args)
	if err := checkBucketName(bucketName); err != nil {
		Error(cmd, args, err)
	}

	// 强制移除桶
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

	// 移除桶
	response, err := sdk.Bucket.RemoveBucket(bucketId, force)
	if err != nil {
		Error(cmd, args, err)
	}

	bucketRemoveRunOutput(cmd, args, response)
}

func bucketRemoveRunOutput(cmd *cobra.Command, args []string, resp model.BucketRemoveResponse) {
	respCode := int(resp.Code)

	if respCode == sdkcode.ErrBucketMustBeEmpty.Code() {
		err := errors.New("Error: Bucket contains objects, add --force to confirm deletion\n")
		Error(cmd, args, err)
	} else if respCode != http.StatusOK {
		Error(cmd, args, errors.New(resp.Msg))
	}

	bucketRemoveOutput := BucketRemoveOutput{
		RequestId: resp.RequestId,
		Code:      resp.Code,
		Msg:       resp.Msg,
		Status:    resp.Status,
	}

	templateContent := `
Succeed
Status: {{.Code}}
`

	t, err := template.New("bucketRemoveTemplate").Parse(templateContent)
	if err != nil {
		Error(cmd, args, err)
	}

	err = t.Execute(os.Stdout, bucketRemoveOutput)
	if err != nil {
		Error(cmd, args, err)
	}
}

type BucketRemoveOutput struct {
	RequestId string       `json:"requestId,omitempty"`
	Code      int32        `json:"code,omitempty"`
	Msg       string       `json:"msg,omitempty"`
	Status    string       `json:"status,omitempty"`
	Data      BucketOutput `json:"bucketOutput,omitempty"`
}

// endregion Bucket Remove

// region Bucket Empty

//var bucketEmptyCmd = &cobra.Command{
//	Use:     "rm",
//	Short:   "rm",
//	Long:    "empty bucket",
//	Example: "gcscmd rm cs://[BUCKET]",
//
//	Run: func(cmd *cobra.Command, args []string) {
//		bucketEmptyRun(cmd, args)
//	},
//}

func bucketEmptyRun(cmd *cobra.Command, args []string) {
	// 桶名称
	bucketName := GetBucketName(args)
	if err := checkBucketName(bucketName); err != nil {
		Error(cmd, args, err)
	}

	// 强制清空桶
	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		Error(cmd, args, err)
	}

	// todo: remove it
	//fmt.Sprint(force)
	if !force {
		Error(cmd, args, errors.New("empty bucket operation, add --force to confirm emptying"))
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

	// 清空桶
	response, err := sdk.Bucket.EmptyBucket(bucketId)
	if err != nil {
		Error(cmd, args, err)
	}

	bucketEmptyRunOutput(cmd, args, response)
}

func bucketEmptyRunOutput(cmd *cobra.Command, args []string, resp model.BucketEmptyResponse) {
	code := resp.Code
	if code != http.StatusOK {
		Error(cmd, args, errors.New(resp.Msg))
	}

	bucketEmptyOutput := BucketEmptyOutput{
		RequestId: resp.RequestId,
		Code:      resp.Code,
		Msg:       resp.Msg,
		Status:    resp.Status,
	}

	templateContent := `
Succeed
Status: {{.Code}}
`

	t, err := template.New("bucketEmptyTemplate").Parse(templateContent)
	if err != nil {
		Error(cmd, args, err)
	}

	err = t.Execute(os.Stdout, bucketEmptyOutput)
	if err != nil {
		Error(cmd, args, err)
	}
}

type BucketEmptyOutput struct {
	RequestId string       `json:"requestId,omitempty"`
	Code      int32        `json:"code,omitempty"`
	Msg       string       `json:"msg,omitempty"`
	Status    string       `json:"status,omitempty"`
	Data      BucketOutput `json:"bucketOutput,omitempty"`
}

// endregion Bucket Empty

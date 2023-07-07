/*
Copyright © 2023 pld

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// rnCmd represents the rn command
var rnCmd = &cobra.Command{
	Use:   "rn cs://BUCKET] [--name=<name>] [--cid=<cid>] [--rename=<rename>] [--force]",
	Short: "rename object",
	Long:  `rename object`,
	Run: func(cmd *cobra.Command, args []string) {
		//fmt.Println("rn called")
		objectRenameRun(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(rnCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// rnCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// rnCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	// 对象名称
	rnCmd.Flags().StringP("name", "n", "", "name of object")

	// 对象CID
	rnCmd.Flags().StringP("cid", "c", "", "cid of object")

	// 重命名
	rnCmd.Flags().StringP("rename", "r", "", "new name of object")

	// 有冲突的时候强制覆盖
	rnCmd.Flags().BoolP("force", "f", false, "if exists filename conflicts, and add --force to confirm operation")
}

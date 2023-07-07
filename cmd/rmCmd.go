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

// rmCmd represents the rm command
var rmCmd = &cobra.Command{
	Use:   "rm cs://[BUCKET] [--name=<name>] [--cid=<cid>] [--force]",
	Short: "empty bucket",
	Long:  `empty bucket`,
	Run: func(cmd *cobra.Command, args []string) {
		//fmt.Println("rm called")

		// 对象名称
		objectName, err := cmd.Flags().GetString("name")
		if err != nil {
			Error(cmd, args, err)
		}

		// 对象CID
		objectCid, err := cmd.Flags().GetString("cid")
		if err != nil {
			Error(cmd, args, err)
		}

		if len(objectName) == 0 && len(objectCid) == 0 {
			bucketEmptyRun(cmd, args)
		} else {
			objectRemoveRun(cmd, args)
		}
	},
}

func init() {
	rootCmd.AddCommand(rmCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// rmCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// rmCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	// 对象名称
	rmCmd.Flags().StringP("name", "n", "", "name of object")

	// 对象CID
	rmCmd.Flags().StringP("cid", "c", "", "cid of object")

	rmCmd.Flags().BoolP("force", "f", false, "confirm emptying bucket")
}

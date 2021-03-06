// Copyright 2015 Nippon Telegraph and Telephone Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"crypto/rand"
	"fmt"
	"os"
	"time"

	"github.com/coreos/etcd/Godeps/_workspace/src/github.com/spf13/cobra"
	"github.com/coreos/etcd/storage"
)

// storagePutCmd represents a storage put performance benchmarking tool
var storagePutCmd = &cobra.Command{
	Use:   "put",
	Short: "Benchmark put performance of storage",

	Run: storagePutFunc,
}

var (
	totalNrKeys    int
	storageKeySize int
	valueSize      int
	txn            bool
)

func init() {
	storageCmd.AddCommand(storagePutCmd)

	storagePutCmd.Flags().IntVar(&totalNrKeys, "total", 100, "a total number of keys to put")
	storagePutCmd.Flags().IntVar(&storageKeySize, "key-size", 64, "a size of key (Byte)")
	storagePutCmd.Flags().IntVar(&valueSize, "value-size", 64, "a size of value (Byte)")
	storagePutCmd.Flags().BoolVar(&txn, "txn", false, "put a key in transaction or not")
}

func createBytesSlice(bytesN, sliceN int) [][]byte {
	rs := make([][]byte, sliceN)
	for i := range rs {
		rs[i] = make([]byte, bytesN)
		if _, err := rand.Read(rs[i]); err != nil {
			panic(err)
		}
	}
	return rs
}

func storagePutFunc(cmd *cobra.Command, args []string) {
	keys := createBytesSlice(storageKeySize, totalNrKeys)
	vals := createBytesSlice(valueSize, totalNrKeys)

	latencies := make([]time.Duration, totalNrKeys)

	minLat := time.Duration(1<<63 - 1)
	maxLat := time.Duration(0)

	for i := 0; i < totalNrKeys; i++ {
		begin := time.Now()

		if txn {
			id := s.TxnBegin()
			if _, err := s.TxnPut(id, keys[i], vals[i], storage.NoLease); err != nil {
				fmt.Errorf("txn put error: %v", err)
				os.Exit(1)
			}
			s.TxnEnd(id)
		} else {
			s.Put(keys[i], vals[i], storage.NoLease)
		}

		end := time.Now()

		lat := end.Sub(begin)
		latencies[i] = lat
		if maxLat < lat {
			maxLat = lat
		}
		if lat < minLat {
			minLat = lat
		}
	}

	total := time.Duration(0)

	for _, lat := range latencies {
		total += lat
	}

	fmt.Printf("total: %v\n", total)
	fmt.Printf("average: %v\n", total/time.Duration(totalNrKeys))
	fmt.Printf("rate: %4.4f\n", float64(totalNrKeys)/total.Seconds())
	fmt.Printf("minimum latency: %v\n", minLat)
	fmt.Printf("maximum latency: %v\n", maxLat)

	// TODO: Currently this benchmark doesn't use the common histogram infrastructure.
	// This is because an accuracy of the infrastructure isn't suitable for measuring
	// performance of kv storage:
	// https://github.com/coreos/etcd/pull/4070#issuecomment-167954149
}

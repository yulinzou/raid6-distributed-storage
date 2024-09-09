package main

import (
    "fmt"
    "raid6-project/raid6"
)

func main() {
    raid := raid6.InitRaid6()

    // Example file data (representing as bytes)
    fileData := []byte("this_is_a_test_file_data")

    // Write file to RAID 6
    err := raid.WriteFile("test.txt", fileData)
    if err != nil {
        fmt.Println(err)
        return
    }

    // Check for corruption
    if raid.CheckCorruption() {
        fmt.Println("Data is corrupted!")
    } else {
        fmt.Println("Data is safe.")
    }

    // Need to edit
    if err := raid.RecoverData(); err != nil {
        fmt.Println(err)
    }
}

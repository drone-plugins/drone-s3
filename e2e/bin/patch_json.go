//usr/bin/env go run "$0" "$@" ; exit "$?"
package main

import (
    "encoding/json"
    "log"
    "os"
)

func mergeMaps(dst map[string]interface{}, src map[string]interface{}) map[string]interface{} {
    for k := range src {
        if src[k] == nil {
            // they want to delete the key
            delete(dst, k)
        } else {
            dstVal, ok := dst[k].(map[string]interface{})
            if ok {
                srcVal, ok := src[k].(map[string]interface{})
                if ok {
                    // both are object - merge
                    dst[k] = mergeMaps(dstVal, srcVal)
                } else {
                    // replace with a primitive
                    dst[k] = src[k]
                }
            } else {
                // does not exist or not an object - replace
                dst[k] = src[k]
            }
        }
    }
    return dst
}

func main() {
    file, err := os.Open(os.Args[1])
    if err != nil {
        log.Println(err)
        return
    }
    defer file.Close()

    decTemplate := json.NewDecoder(file)
    decDelta := json.NewDecoder(os.Stdin)
    enc := json.NewEncoder(os.Stdout)

    var template, delta map[string]interface{}
    if err := decTemplate.Decode(&template); err != nil {
        log.Println(err)
        return
    }
    if err := decDelta.Decode(&delta); err != nil {
        log.Println(err)
        return
    }

    dst := mergeMaps(template, delta)
    if err := enc.Encode(&dst); err != nil {
        log.Println(err)
    }
}

## rust_strip

A simple tool to remove unused rust use imports.


1. Cargo b
2. Read the warning logs
3. Replace the files


## usage

```
go get -u -v github.com/sundy-li/rust_strip

cd your_sub_crate_path
rust_strip -root=you_root_crate_path

## for test files
rust_strip -root=you_root_crate_path -test
```

## More
```
rust_strip -h
```
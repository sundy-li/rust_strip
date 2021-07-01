## rust_strip

A simple tool to remove unused imports in RUST.


1. Cargo build/tests
2. Process the warning logs of unused imports
3. Then replace the involved files


## usage

```
go get -u -v github.com/sundy-li/rust_strip

cd your_crate_path

rust_strip -root=your_root_crate_path

## for test files
rust_strip -root=your_root_crate_path -test
```

## More
```
rust_strip -h
```
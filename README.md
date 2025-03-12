## Channel Check 

Channels are a great feature of Golang but have several footguns that can lead to deadlocks. In particular, if the receiving channel stops processing the messages, a *non-blocking* channel send would fail to continue. In certain mission-critical sections of code, this could lead to a complete deadlock. 
  
This linter currently has three features: 
- Non-blocking sends 
- Non-buffered channel creation detection 
- Buffered channel size exceeds maximum size checks 
  
Many of these will lead to false positives or situations where we *want* a blocking channel send. In these cases, `nolint:channelcheck` is easy to add. Regardless, having this issue pointed out automatically is a good way to fix bugs; this doesn't necessarily have to be included in CI. 

## Golangci-lint Integration 
This is the recommended way to use the linter. golangci-lint has a [module plugin](https://golangci-lint.run/plugins/module-plugins/) system that works very well. To install the linter this way, do the following: 

1. Download this repo locally.
2. In your target repo, create a `.custom-gcl.yml` file. Add the following content: 
   NOTE: This can't be directly copied. Use the contents of `.example-custom-gcl.yml`.
    ```yaml
        version: v1.64.5
        plugins:
        # a plugin from local source
        - module: github.com/asymmetric-research/channel_linter/channelcheck
            path: /Path/to/channelcheck_repo
        name: golangci-lint-with-channelcheck
    ```
3. In your `.golangci.yaml` file, add the following and enable the linter. 
  NOTE: This can't be directly copied. Use the contents of `.example-golangci.yml`.
    ```yaml 
    linters-settings:
        custom: 
            channelcheck:
            type: "module" 
            description: Static analysis for go channel issues   
            settings: 
                CheckBlockingSends: true
                CheckUnbufferedChannels: false
    ```
4. Build the custom version of `golangci-lint`. This is literally recompiling the linter binary and adding our linter into it.
    ```bash 
        golangci-lint custom -v
    ```
5. Run the new binary on your repo: 
    ```bash
        ./golangci-lint-with-channelcheck run 
    ```
6. To remove false positives, add `nolint:channelcheck` above the line that had the linter error.

## Configuration Steps Standalone 
The linter can be used by itself. Simply run the following to install the binary: 

```bash
go install ./cmd/channellint/main.go
```

Usage: 

```bash 
channellint ./examples
```

This is NOT recommended because false positives cannot be tuned out via `nolint` comments.



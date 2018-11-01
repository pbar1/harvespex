# Harvespex

As of right now, this is what it does:
1. Find dates in need of Harvest entries
2. Fetch, filter, and aggregate  GitHub commit messages for those dates
3. Print them out for easy copypasta :stuck_out_tongue_closed_eyes:

To run, simply: `./harvespex` or `go run main.go`

### Setup
You'll need some environment variables:
- `GITHUB_TOKEN`
- `HARVEST_ACCESS_TOKEN`
- `HARVEST_ACCOUNT_ID`

...and you'll also need `harvespex.hcl` in the same directory as the `harvespex` binary. This maps GitHub repositories to Harvest tasks, and looks like so:

```
project_mapping {
  project = "My Exact Harvest Project Name"
  task = "My Exact Harvest Task Name"
  repositories = [
    "pbar1/harvespex",
    "myorg/example-proj",
  ]
}
```


# Blog RSS aggregator - gator 
Simple RSS blog data aggregator (gator)

Uses ~/gatorconfig.json to store database connection settings and current user login


# Installation
Required postgres in versio >= 14 and golang tool chain >= 1.23.4.
In main folder run `go build -o gator` and then `go install`. 
Run commands with `./gator login <username>`

# Configuration

Create config in main path ~/gatorconfig.json. File consists of database url and current signed in json. 

```json
{
 "db_url": "postgres://postgres:postgres@localhost:5432/gator?sslmode=disable",
 "current_user_name": "unknown"
}
```

# Example commands
`register <name>` -> adds new user to database
`addfeed <name> <feed url>` -> Add new feed source to program
`agg <time_interval>` - eg. agg 30s every 30s RSS feeds will be aggregated to program
`browse <num_of_posts>` - browse through articles titles
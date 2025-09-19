# aggre(GATOR)

Gator is a simple CLI rss feed aggregator using Postgres and Go. To use you will need Postgres and Go installed on a current linux distro.  

## Install
To compile and install use
```bash
go install github.com/jdfincher/gator
```

## Setup
In your home directory create a file called `.gatorconfig.json` and include the below information. This is for the default postgres connection string with the default postgres user. If in setting up your postgres server you change the user name or db name you will need to updated the `"db_url":` field accordingly. Leave the `"current_user_name":""` as is. 

```json
{
"db_url":"postgres://postgres:@localhost:5432/gator?sslmode=disable",
"current_user_name":""
}
 // db_url formatting is protocol://username:password@host:port/database_name
```

If you don't already have Go and PostgreSQL installed do so now.
 >[Go Install](https://go.dev/doc/install)
 >[PostgreSQL install](https://www.postgresql.org/download/)

Initialize the data cluster if this is a new postgres install, check paths/cmds for your distro
```bash
sudo -iu postgres #this is the default user for postgres
init -D /var/lib/postgres/data #initialize data cluster
exit #logout of postgres user
```

Start postgres
```bash
sudo systemctl start postgresql #starts server
sudo systemctl status postgresql #verify that it is running
```

Login to the psql shell and create the gator database
```bash
sudo -iu postgres psql #switchs users and logs into psql shell as user postgres
CREATE DATABASE gator; #create database
\c gator #connect to gator database
ALTER USER postgres PASSWORD 'new' #change postgres password 
```

Now I would recommend installing goose to handle creating and migrating the tables in the database. Otherwise you will need to manually enter each migration found in sql/schema into the psql shell. 
```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

Create a directory wherever you want to store the database schema info, this is where you will place all the migrations found at the top in sql/schema ie "001_users.sql". From the schema folder run the following command with your connection string.
```bash
goose postgres://postgres:@localhost:5432/gator up
```

This should report back it successfully migrated to `version: 5` you can check that the database is setup correctly by logging back into the psql shell and checking the tables. 
```bash
sudo -iu postgres psql gator
\dt
```

If everything is setup properly you should see the below
```bash
List of relations
 Schema |       Name       | Type  |  Owner
--------+------------------+-------+----------
 public | feed_follows     | table | postgres
 public | feeds            | table | postgres
 public | goose_db_version | table | postgres
 public | posts            | table | postgres
 public | users            | table | postgres
(5 rows)
```

## Commands
First register yourself as a user in the db. This will register your username and log you in.
```bash
./gator register "user name"
```

To change currently logged in user.
```bash
./gator login "user name"
```

List all users and show current.
```bash
./gator users
```

Add feed url. This will also automatically follow the feed for the logged in user.  
```bash
./gator addfeed "name" "url"
```

View all feeds added by all users.
```bash
./gator feeds
```

Follow and Unfollow feeds.
```bash
./gator follow "url"
./gator unfollow "url"
```

Fetch all posts from feed urls in continuous loop with the time interval you set. Time intervals are in the format "#h#m#s" for example "30s" for 30 seconds. 
```bash
./gator agg "interval"
```

Finally browse posts sorted by published date with an optional "limit" argument to limit the amount of posts displayed at a time, the default is 2 if no argument is passed. 
```bash
./gator browse "limit"   #Returns 2 if limit amount omitted
./gator browse 100 | less #Might want to pipe to a pager if viewing many
```

Reset the state of the database with the reset command. *Warning this wipes the entire database in an unrecoverable way, use with caution!*
```bash
./gator reset #Returns database to fresh install state
```

![gatorCLI](img/gator.png)

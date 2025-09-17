package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jdfincher/gator/internal/config"
	"github.com/jdfincher/gator/internal/database"
	_ "github.com/lib/pq"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	commands map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	f := c.commands[cmd.name]
	err := f(s, cmd)
	if err != nil {
		return fmt.Errorf("error: running %v command -> %w", cmd.name, err)
	}
	return nil
}

func (c *commands) register(name string, f func(*state, command) error) error {
	c.commands[name] = f
	return nil
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("error: login requires Username")
	}
	_, err := s.db.GetUser(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("error: user %v does not exist, use <register %v> to create user -> %w", cmd.args[0], cmd.args[0], err)
	}
	if err := s.cfg.SetUser(cmd.args[0]); err != nil {
		return err
	}
	fmt.Printf("Username successfully set to %v\n", cmd.args[0])
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("error: register requires Username")
	}
	newUser := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.args[0],
	}
	user, err := s.db.CreateUser(context.Background(), newUser)
	if err != nil {
		return fmt.Errorf("error: user could not be created -> %w", err)
	}
	if err := handlerLogin(s, cmd); err != nil {
		return err
	}
	fmt.Println(strings.Repeat("*", 20))
	fmt.Printf("User ID -> %v\n", user.ID)
	fmt.Printf("Created At -> %v\n", user.CreatedAt)
	fmt.Printf("Updated At -> %v\n", user.UpdatedAt)
	fmt.Printf("Name -> %v\n", user.Name)
	fmt.Println(strings.Repeat("*", 20))
	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.Reset(context.Background())
	if err != nil {
		return fmt.Errorf("error: issue reseting users table -> %w", err)
	}
	fmt.Printf("users table was successfully cleared\n")
	return nil
}

func handlerUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error: issue retriever user table records -> %w", err)
	}
	for i := range users {
		if users[i].Name == s.cfg.UserName {
			fmt.Printf("* %v (current)\n", users[i].Name)
			continue
		}
		fmt.Printf("* %v\n", users[i].Name)
	}
	return nil
}

func handlerAgg(s *state, cmd command) error {
	feed, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return err
	}
	fmt.Println(feed)
	return nil
}

func handlerAddFeed(s *state, cmd command) error {
	if len(cmd.args) < 2 {
		return fmt.Errorf("error: cannot add feed to database without name & url")
	}
	user, err := s.db.GetUser(context.Background(), s.cfg.UserName)
	if err != nil {
		return fmt.Errorf("error: cannot find user record -> %w", err)
	}
	newfeed := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.args[0],
		Url:       cmd.args[1],
		UserID:    user.ID,
	}
	feed, err := s.db.CreateFeed(context.Background(), newfeed)
	if err != nil {
		return fmt.Errorf("error: cannot create feed record -> %w", err)
	}
	fmt.Println(feed)
	return nil
}

func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("error: could not fetch feeds, maybe there are none? -> %w", err)
	}
	if feeds == nil {
		return fmt.Errorf("error: there are no feeds to show! -> %w", err)
	}
	for i := range feeds {
		name, err := s.db.GetUserByID(context.Background(), feeds[i].UserID)
		if err != nil {
			return fmt.Errorf("error: issue retrieving user name for feed record -> %w", err)
		}
		num := i + 1
		fmt.Printf("----|Feed - %d|----\n", num)
		fmt.Printf("%v\n", feeds[i].Name)
		fmt.Printf("%v\n", feeds[i].Url)
		fmt.Printf("%v\n", name)
	}
	fmt.Printf("---|End of list|---\n")
	return nil
}

func newCommands() *commands {
	c := new(commands)
	c.commands = make(map[string]func(*state, command) error)
	return c
}

func main() {
	var err error
	state := new(state)
	state.cfg, err = config.Read()
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	db, err := sql.Open("postgres", state.cfg.DBURL)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	state.db = database.New(db)
	coms := newCommands()
	if err = coms.register("login", handlerLogin); err != nil {
		fmt.Printf("%v\n", err)
	}
	if err = coms.register("register", handlerRegister); err != nil {
		fmt.Printf("%v\n", err)
	}
	if err = coms.register("reset", handlerReset); err != nil {
		fmt.Printf("%v\n", err)
	}
	if err = coms.register("users", handlerUsers); err != nil {
		fmt.Printf("%v\n", err)
	}
	if err = coms.register("agg", handlerAgg); err != nil {
		fmt.Printf("%v\n", err)
	}
	if err = coms.register("addfeed", handlerAddFeed); err != nil {
		fmt.Printf("%v\n", err)
	}
	if err = coms.register("feeds", handlerFeeds); err != nil {
		fmt.Printf("%v\n", err)
	}
	args := os.Args
	if len(args) < 2 {
		fmt.Printf("error: missing command arguments, use gator 'command_name'\n")
		os.Exit(1)
	}
	usercom := command{
		name: strings.ToLower(args[1]),
		args: args[2:],
	}
	if err := coms.run(state, usercom); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}

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

func (c *commands) register(name string, f func(*state, command) error) {
	c.commands[name] = f
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
	fmt.Printf(`
░█░░░█▀█░█▀▀░█▀▀░█▀▀░█▀▄░░░▀█▀░█▀█
░█░░░█░█░█░█░█░█░█▀▀░█░█░░░░█░░█░█
░▀▀▀░▀▀▀░▀▀▀░▀▀▀░▀▀▀░▀▀░░░░▀▀▀░▀░▀
`+"---> %v\n", cmd.args[0])
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

	fmt.Printf(`
░█▀▄░█▀▀░█▀▀░▀█▀░█▀▀░▀█▀░█▀▀░█▀▄░█▀▀░█▀▄
░█▀▄░█▀▀░█░█░░█░░▀▀█░░█░░█▀▀░█▀▄░█▀▀░█░█
░▀░▀░▀▀▀░▀▀▀░▀▀▀░▀▀▀░░▀░░▀▀▀░▀░▀░▀▀▀░▀▀░
		` + "\n")
	fmt.Println(strings.Repeat("*", 52))
	fmt.Printf("User ID -> %v\n", user.ID)
	fmt.Printf("Created At -> %v\n", user.CreatedAt)
	fmt.Printf("Updated At -> %v\n", user.UpdatedAt)
	fmt.Printf("Name -> %v\n", user.Name)
	fmt.Println(strings.Repeat("*", 52))
	if err := handlerLogin(s, cmd); err != nil {
		return err
	}
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
	fmt.Printf(`
░█░█░█▀▀░█▀▀░█▀▄░█▀▀
░█░█░▀▀█░█▀▀░█▀▄░▀▀█
░▀▀▀░▀▀▀░▀▀▀░▀░▀░▀▀▀` + "\n")
	for i := range users {
		if users[i].Name == s.cfg.UserName {
			fmt.Printf("> %v (current)\n", users[i].Name)
			continue
		}
		fmt.Printf("> %v\n", users[i].Name)
	}
	return nil
}

func handlerAgg(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("error: time between requests required use 'agg 5s' to set interval to 5 seconds")
	}
	reqInterval, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	ticker := time.NewTicker(reqInterval)
	defer ticker.Stop()
	fmt.Printf("Collecting feeds every %v\n", reqInterval)
	for range ticker.C {
		err := scrapeFeeds(s)
		if err != nil {
			return err
		}
	}
	return nil
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 2 {
		return fmt.Errorf("error: could not add feed to database without name & url")
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
		return fmt.Errorf("error: could not create feed record -> %w", err)
	}
	autoFollow := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	}
	_, err = s.db.CreateFeedFollow(context.Background(), autoFollow)
	if err != nil {
		return fmt.Errorf("error: could not follow created feed -> %w", err)
	}
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

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("error: what are you trying to follow? Use 'follow url' to follow")
	}
	feedID, err := s.db.GetFeedID(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("error: issue fetching feed ID from database -> %w", err)
	}
	follow := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feedID,
	}
	cF, err := s.db.CreateFeedFollow(context.Background(), follow)
	if err != nil {
		return fmt.Errorf("error: could not create feed follow record -> %w", err)
	}
	fmt.Printf("User: %v followed -> '%v'\n", cF.UserName, cF.FeedName)
	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	follows, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("error: issue fetching follows for user from database -> %w", err)
	}
	tot := len(follows)
	for i := range follows {
		num := i + 1
		fmt.Printf("---|%v of %v|---\n", num, tot)
		fmt.Printf("%v\n", follows[i].FeedName)
	}
	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("error: no url to unfollow provided try unfollow 'url'")
	}
	feedID, err := s.db.GetFeedID(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("error: could not fetch feed id from database based on provided url -> %w", err)
	}
	deleteFeed := database.DeleteFeedFollowForUserParams{
		UserID: user.ID,
		FeedID: feedID,
	}
	if err := s.db.DeleteFeedFollowForUser(context.Background(), deleteFeed); err != nil {
		return fmt.Errorf("error: failed to cleanly remove follow record")
	}
	fmt.Printf("Successfully unfollowed %v\n", cmd.args[0])
	return nil
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		user, err := s.db.GetUser(context.Background(), s.cfg.UserName)
		if err != nil {
			return fmt.Errorf("error: user must be logged in, try login 'username' or register 'username'")
		}
		return handler(s, cmd, user)
	}
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
	coms.register("login", handlerLogin)
	coms.register("register", handlerRegister)
	coms.register("reset", handlerReset)
	coms.register("users", handlerUsers)
	coms.register("agg", middlewareLoggedIn(handlerAgg))
	coms.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	coms.register("feeds", handlerFeeds)
	coms.register("follow", middlewareLoggedIn(handlerFollow))
	coms.register("following", middlewareLoggedIn(handlerFollowing))
	coms.register("unfollow", middlewareLoggedIn(handlerUnfollow))

	args := os.Args
	if len(args) < 2 {
		fmt.Printf("error: missing command arguments, use gator 'command_name'\n")
		os.Exit(1)
	}
	usercom := command{
		name: strings.ToLower(args[1]),
		args: args[2:],
	}
	_, ok := coms.commands[usercom.name]
	if !ok {
		fmt.Printf("error: unkown command '%v'\n", usercom.name)
		os.Exit(1)
	}
	if err := coms.run(state, usercom); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}

package main

import (
    "strings"
    "testing"

    "github.com/bwmarrin/discordgo"
)

func makeInteraction(eventId int64, round int64, pts *float64) *discordgo.Interaction {
    opts := []*discordgo.ApplicationCommandInteractionDataOption{
        {
            Name: string(TdByeCmd),
            Type: discordgo.ApplicationCommandOptionSubCommand,
        },
    }

    // Build the nested options slice for the sub-command
    subOpts := []*discordgo.ApplicationCommandInteractionDataOption{
        {
            Name:  "eventid",
            Type:  discordgo.ApplicationCommandOptionInteger,
            Value: float64(eventId),
        },
        {
            Name:  "round",
            Type:  discordgo.ApplicationCommandOptionInteger,
            Value: float64(round),
        },
    }
    if pts != nil {
        subOpts = append(subOpts, &discordgo.ApplicationCommandInteractionDataOption{
            Name:  "pts",
            Type:  discordgo.ApplicationCommandOptionNumber,
            Value: *pts,
        })
    }
    opts[0].Options = subOpts

        val := discordgo.ApplicationCommandInteractionData{
        Name:    string(TdCmd),
        Options: opts,
    }

    return &discordgo.Interaction{
        Member: &discordgo.Member{User: &discordgo.User{Username: "TestUser"}}, 
        Type: discordgo.InteractionApplicationCommand,
        Data: discordgo.InteractionData(val),
            
    }
}

func TestTdByeCmdHandlerValidation(t *testing.T) {
    // stub tournament fetcher
    fetchTournament = func(id int64) (*Tournament, error) {
        return &Tournament{Players: []Player{{DisplayName: "Test User", UscfID: 42}}}, nil
    }

    // invalid round
    inter := makeInteraction(1, 0, nil)
    resp := tdByeCmdHandler(inter)
    if resp.Data == nil || !strings.HasPrefix(resp.Data.Content, "❌") {
        t.Errorf("expected validation error for round < 1, got %+v", resp.Data)
    }

    // invalid pts
    badPts := 2.0
    inter = makeInteraction(1, 1, &badPts) // eventId=1 round=1
    resp = tdByeCmdHandler(inter)
    if resp.Data == nil || !strings.HasPrefix(resp.Data.Content, "❌") {
        t.Errorf("expected validation error for bad pts, got %+v", resp.Data)
    }

    // valid
    goodPts := 0.5
    inter = makeInteraction(1, 3, &goodPts)
    resp = tdByeCmdHandler(inter)
    if resp.Data == nil || !strings.HasPrefix(resp.Data.Content, "✅") {
        t.Errorf("expected success message, got %+v", resp.Data)
    }
}

package main

import (
    "strings"
    "testing"

    "github.com/bwmarrin/discordgo"
)

func makeInteraction(round int64, pts *float64) *discordgo.Interaction {
    opts := []*discordgo.ApplicationCommandInteractionDataOption{
        {
            Name: string(TdByeCmd),
            Type: discordgo.ApplicationCommandOptionSubCommand,
        },
    }

    // Build the nested options slice for the sub-command
    subOpts := []*discordgo.ApplicationCommandInteractionDataOption{
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
        Type: discordgo.InteractionApplicationCommand,
        Data: discordgo.InteractionData(val),
            
    }
}

func TestTdByeCmdHandlerValidation(t *testing.T) {
    // invalid round
    inter := makeInteraction(0, nil)
    resp := tdByeCmdHandler(inter)
    if resp.Data == nil || !strings.HasPrefix(resp.Data.Content, "❌") {
        t.Errorf("expected validation error for round < 1, got %+v", resp.Data)
    }

    // invalid pts
    badPts := 2.0
    inter = makeInteraction(1, &badPts)
    resp = tdByeCmdHandler(inter)
    if resp.Data == nil || !strings.HasPrefix(resp.Data.Content, "❌") {
        t.Errorf("expected validation error for bad pts, got %+v", resp.Data)
    }

    // valid
    goodPts := 0.5
    inter = makeInteraction(3, &goodPts)
    resp = tdByeCmdHandler(inter)
    if resp.Data == nil || !strings.HasPrefix(resp.Data.Content, "✅") {
        t.Errorf("expected success message, got %+v", resp.Data)
    }
}

package game

import (
	"fmt"
	"strings"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/glog"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/session"
)

func (m *WorldMode) sendGuildInvite(ctx client.Context, actorID uint32, name string) {
	name = strings.TrimSpace(name)
	if actorID == 0 {
		return
	}
	if ctx.Network == nil {
		glog.Warnf("guild invite failed target=%d name=%q: not connected", actorID, name)
		m.ui.console.AddErrorMessage("Guild invitation failed: not connected.")
		return
	}
	accountID, charID := uint32(0), uint32(0)
	if ctx.Session != nil {
		accountID = ctx.Session.AccountID
		charID = ctx.Session.CharID
	}
	if err := ctx.Network.SendGuildInvite(actorID, accountID, charID); err != nil {
		glog.Warnf("guild invite failed target=%d name=%q: %v", actorID, name, err)
		m.ui.console.AddErrorMessage("Guild invitation failed.")
		return
	}
	m.ui.console.AddBlueMessage("%s has received an invitation to join your guild.", guildDisplayName(name))
}

func (m *WorldMode) openGuildInviteRequest(ctx client.Context, request network.GuildInviteRequest) {
	name := guildDisplayName(request.GuildName)
	rawName := strings.TrimSpace(request.GuildName)
	m.ui.guildRequest.Open(ctx, "Guild Invitation", fmt.Sprintf("Would you like to join %s?", name), func() {
		if ctx.Network == nil {
			glog.Warnf("guild invite accept failed: not connected")
			return
		}
		if err := ctx.Network.SendGuildInviteReply(request.GuildID, true); err != nil {
			glog.Warnf("guild invite accept failed guild=%d name=%q: %v", request.GuildID, request.GuildName, err)
			return
		}
		applyLocalGuildName(ctx, rawName)
	}, func() {
		if ctx.Network == nil {
			glog.Warnf("guild invite reject failed: not connected")
			return
		}
		if err := ctx.Network.SendGuildInviteReply(request.GuildID, false); err != nil {
			glog.Warnf("guild invite reject failed guild=%d name=%q: %v", request.GuildID, request.GuildName, err)
		}
	})
}

func (m *WorldMode) handleGuildCreationResult(ctx client.Context, result network.GuildCreationResult) {
	switch result.Result {
	case 0:
		if name := pendingGuildName(ctx); name != "" {
			applyLocalGuildName(ctx, name)
		}
		m.ui.console.AddBlueMessage("Guild created.")
	case 1:
		clearPendingGuildName(ctx)
		m.ui.console.AddErrorMessage("You are already in a guild.")
	case 2:
		clearPendingGuildName(ctx)
		m.ui.console.AddErrorMessage("Guild name already exists.")
	case 3:
		clearPendingGuildName(ctx)
		m.ui.console.AddErrorMessage("You need the required item to create a guild.")
	default:
		clearPendingGuildName(ctx)
		m.ui.console.AddErrorMessage("Guild creation failed.")
	}
}

func (m *WorldMode) handleGuildInviteAck(ack network.GuildInviteAck) {
	switch ack.Result {
	case 0:
		m.ui.console.AddErrorMessage("That character is already in a guild.")
	case 1:
		m.ui.console.AddErrorMessage("Guild invitation was rejected.")
	case 2:
		m.ui.console.AddBlueMessage("Guild invitation accepted.")
	case 3:
		m.ui.console.AddErrorMessage("The guild is full.")
	default:
		m.ui.console.AddErrorMessage("Guild invitation failed.")
	}
}

func guildDisplayName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "Guild"
	}
	return name
}

func pendingGuildName(ctx client.Context) string {
	if ctx.Session == nil {
		return ""
	}
	return strings.TrimSpace(ctx.Session.PendingGuildName)
}

func clearPendingGuildName(ctx client.Context) {
	if ctx.Session != nil {
		ctx.Session.PendingGuildName = ""
	}
}

func applyLocalGuildName(ctx client.Context, name string) {
	applyLocalGuildInfo(ctx, 0, 0, name)
}

func applyLocalGuildDetails(ctx client.Context, info network.GuildInfo) {
	applyLocalGuildInfo(ctx, info.GuildID, info.EmblemVersion, info.GuildName)
	if ctx.Session != nil {
		members := ctx.Session.Guild.Members
		positions := ctx.Session.Guild.Positions
		skillPoints := ctx.Session.Guild.SkillPoints
		skills := ctx.Session.Guild.Skills
		ctx.Session.Guild = session.Guild{
			ID:               info.GuildID,
			Level:            info.Level,
			UserNum:          info.UserNum,
			MaxUserNum:       info.MaxUserNum,
			UserAverageLevel: info.UserAverageLevel,
			Exp:              info.Exp,
			MaxExp:           info.MaxExp,
			Point:            info.Point,
			Honor:            info.Honor,
			Virtue:           info.Virtue,
			EmblemVersion:    info.EmblemVersion,
			Name:             strings.TrimSpace(info.GuildName),
			MasterName:       strings.TrimSpace(info.MasterName),
			ManageLand:       strings.TrimSpace(info.ManageLand),
			Zeny:             info.Zeny,
			Members:          members,
			Positions:        positions,
			SkillPoints:      skillPoints,
			Skills:           skills,
		}
	}
}

func applyLocalGuildMembers(ctx client.Context, members []network.GuildMember) {
	if ctx.Session == nil {
		return
	}
	ctx.Session.Guild.Members = make([]session.GuildMember, 0, len(members))
	var online uint32
	for _, member := range members {
		if member.CurrentState != 0 {
			online++
		}
		ctx.Session.Guild.Members = append(ctx.Session.Guild.Members, session.GuildMember{
			AccountID:    member.AccountID,
			CharID:       member.CharID,
			HeadType:     member.HeadType,
			HeadPalette:  member.HeadPalette,
			Sex:          member.Sex,
			Job:          member.Job,
			Level:        member.Level,
			MemberExp:    member.MemberExp,
			CurrentState: member.CurrentState,
			PositionID:   member.PositionID,
			Memo:         strings.TrimSpace(member.Memo),
			CharName:     strings.TrimSpace(member.CharName),
		})
	}
	ctx.Session.Guild.UserNum = uint32(len(members))
	glog.Debugf("guild member list received members=%d online=%d", len(members), online)
}

func applyLocalGuildPositions(ctx client.Context, positions []network.GuildPosition) {
	if ctx.Session == nil {
		return
	}
	for _, position := range positions {
		index := guildPositionIndex(ctx.Session.Guild.Positions, position.PositionID)
		if index < 0 {
			ctx.Session.Guild.Positions = append(ctx.Session.Guild.Positions, session.GuildPosition{
				PositionID: position.PositionID,
				Right:      position.Right,
				Ranking:    position.Ranking,
				PayRate:    position.PayRate,
			})
			continue
		}
		ctx.Session.Guild.Positions[index].Right = position.Right
		ctx.Session.Guild.Positions[index].Ranking = position.Ranking
		ctx.Session.Guild.Positions[index].PayRate = position.PayRate
	}
	glog.Debugf("guild positions received positions=%d", len(positions))
}

func applyLocalGuildPositionNames(ctx client.Context, positions []network.GuildPosition) {
	if ctx.Session == nil {
		return
	}
	for _, position := range positions {
		index := guildPositionIndex(ctx.Session.Guild.Positions, position.PositionID)
		if index < 0 {
			ctx.Session.Guild.Positions = append(ctx.Session.Guild.Positions, session.GuildPosition{
				PositionID: position.PositionID,
				PosName:    strings.TrimSpace(position.PosName),
			})
			continue
		}
		ctx.Session.Guild.Positions[index].PosName = strings.TrimSpace(position.PosName)
	}
	glog.Debugf("guild position names received positions=%d", len(positions))
}

func applyLocalGuildSkills(ctx client.Context, info network.GuildSkillInfo) {
	if ctx.Session == nil {
		return
	}
	ctx.Session.Guild.SkillPoints = info.SkillPoints
	ctx.Session.Guild.Skills = make([]session.Skill, 0, len(info.Skills))
	for _, skill := range info.Skills {
		ctx.Session.Guild.Skills = append(ctx.Session.Guild.Skills, sessionSkillFromNetworkWithResources(ctx.Resources, skill))
	}
	glog.Debugf("guild skill list received count=%d points=%d", len(info.Skills), info.SkillPoints)
}

func guildPositionIndex(positions []session.GuildPosition, id uint32) int {
	for i, position := range positions {
		if position.PositionID == id {
			return i
		}
	}
	return -1
}

func applyLocalGuildInfo(ctx client.Context, guildID, emblemVersion uint32, name string) {
	name = strings.TrimSpace(name)
	if name == "" && guildID == 0 && emblemVersion == 0 {
		clearPendingGuildName(ctx)
		return
	}
	if ctx.Session != nil {
		if guildID != 0 {
			ctx.Session.GuildID = guildID
		}
		if emblemVersion != 0 {
			ctx.Session.EmblemVersion = emblemVersion
		}
		if name != "" {
			ctx.Session.GuildName = name
		}
		if guildID != 0 {
			ctx.Session.Guild.ID = guildID
		}
		if emblemVersion != 0 {
			ctx.Session.Guild.EmblemVersion = emblemVersion
		}
		if name != "" {
			ctx.Session.Guild.Name = name
		}
		ctx.Session.PendingGuildName = ""
	}
	if ctx.World != nil {
		if guildID != 0 {
			ctx.World.Player.GuildID = guildID
		}
		if emblemVersion != 0 {
			ctx.World.Player.EmblemVersion = emblemVersion
		}
		if name != "" {
			ctx.World.Player.GuildName = name
		}
	}
}

package pwengine

import (
	"context"
	"fmt"
	"time"

	"pathwar.land/go/pkg/pwdb"
)

func (e *engine) UserDeleteAccount(ctx context.Context, in *UserDeleteAccount_Input) (*UserDeleteAccount_Output, error) {
	userID, err := userIDFromContext(ctx, e.db)
	if err != nil {
		return nil, fmt.Errorf("get userid from context: %w", err)
	}
	now := time.Now()

	// get user
	var user pwdb.User
	err = e.db.
		Preload("TeamMemberships").
		Preload("TeamMemberships.Team.Members.User").
		Preload("OrganizationMemberships").
		Preload("OrganizationMemberships.Organization.Members.User").
		First(&user, userID).
		Error
	if err != nil {
		return nil, err
	}

	//fmt.Println(godev.PrettyJSON(user))

	// update user
	updates := pwdb.User{
		OAuthSubject:   fmt.Sprintf("deleted_%s_%d", user.OAuthSubject, now.Unix()),
		DeletionReason: in.Reason,
		DeletionStatus: pwdb.DeletionStatus_Requested,
		DeletedAt:      &now,
	}
	err = e.db.Model(&user).Updates(updates).Error
	if err != nil {
		return nil, err
	}

	// update teams
	for _, teamMembership := range user.TeamMemberships {
		haveAnotherActiveMember := false
		for _, member := range teamMembership.Team.Members {
			if member.User.ID == user.ID {
				continue
			}
			if member.User.DeletionStatus == pwdb.DeletionStatus_Active {
				haveAnotherActiveMember = true
				break
			}
		}
		if !haveAnotherActiveMember {
			updates := pwdb.Team{
				DeletionStatus: pwdb.DeletionStatus_Requested,
				DeletedAt:      &now,
			}
			err = e.db.Model(&teamMembership.Team).Updates(updates).Error
			if err != nil {
				return nil, err
			}
		}
	}

	// update organizations
	for _, organizationMembership := range user.OrganizationMemberships {
		haveAnotherActiveMember := false
		for _, member := range organizationMembership.Organization.Members {
			if member.User.ID == user.ID {
				continue
			}
			if member.User.DeletionStatus == pwdb.DeletionStatus_Active {
				haveAnotherActiveMember = true
				break
			}
		}
		if !haveAnotherActiveMember {
			updates := pwdb.Organization{
				Name:           fmt.Sprintf("deleted_%s_%d", organizationMembership.Organization.Name, now.Unix()),
				DeletionStatus: pwdb.DeletionStatus_Requested,
				DeletedAt:      &now,
			}
			err = e.db.Model(&organizationMembership.Organization).Updates(updates).Error
			if err != nil {
				return nil, err
			}
		}
	}

	// FIXME: invalide current JWT token
	// FIXME: add another task that pseudonymize the data

	ret := &UserDeleteAccount_Output{}
	return ret, nil
}
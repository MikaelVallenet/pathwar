package pwapi

import (
	"context"

	"github.com/jinzhu/gorm"

	"pathwar.land/pathwar/v2/go/pkg/errcode"
	"pathwar.land/pathwar/v2/go/pkg/pwdb"
)

func (svc *service) OrganizationSendInvite(ctx context.Context, in *OrganizationSendInvite_Input) (*OrganizationSendInvite_Output, error) {
	if in == nil || in.OrganizationID == "" || in.UserID == "" {
		return nil, errcode.ErrMissingInput
	}

	userID, err := userIDFromContext(ctx, svc.db)
	if err != nil {
		return nil, errcode.ErrUnauthenticated.Wrap(err)
	}

	organizationID, err := pwdb.GetIDBySlugAndKind(svc.db, in.OrganizationID, "organization")
	if err != nil {
		return nil, err
	}

	inviteUserID, err := pwdb.GetIDBySlugAndKind(svc.db, in.UserID, "user")
	if err != nil {
		return nil, err
	}

	// check organization status
	var organization pwdb.Organization
	err = svc.db.
		Where(pwdb.Organization{
			ID:             organizationID,
			DeletionStatus: pwdb.DeletionStatus_Active,
		}).
		First(&organization).
		Error
	if err != nil {
		return nil, errcode.ErrOrganizationDoesNotExist.Wrap(err)
	}

	// check that the user is owner of the organization
	var organizationOwner pwdb.OrganizationMember
	err = svc.db.
		Where(pwdb.OrganizationMember{
			UserID:         userID,
			OrganizationID: organizationID,
			Role:           pwdb.OrganizationMember_Owner,
		}).
		First(&organizationOwner).
		Error
	if err != nil {
		return nil, errcode.ErrNotOrganizationOwner.Wrap(err)
	}

	// check if invited user already is a member of this organization
	var organizationMembership int
	err = svc.db.
		Model(&pwdb.OrganizationMember{}).
		Where(pwdb.OrganizationMember{
			UserID:         inviteUserID,
			OrganizationID: organizationID,
		}).
		Count(&organizationMembership).
		Error
	if err != nil || organizationMembership != 0 {
		return nil, errcode.ErrOrganizationUserAlreadyMember.Wrap(err)
	}

	// don't create a new invite if one already exists
	var organizationInvite pwdb.OrganizationInvite
	err = svc.db.
		Where(pwdb.OrganizationInvite{
			UserID:         inviteUserID,
			OrganizationID: organizationID,
		}).
		First(&organizationInvite).
		Error
	if err == nil {
		return nil, errcode.ErrAlreadyInvitedInOrganization.Wrap(err)
	} else if err != gorm.ErrRecordNotFound {
		return nil, pwdb.GormToErrcode(err)
	}

	organizationInvite = pwdb.OrganizationInvite{
		UserID:         inviteUserID,
		OrganizationID: organizationID,
	}

	err = svc.db.Transaction(func(tx *gorm.DB) error {
		err = tx.Create(&organizationInvite).Error
		if err != nil {
			return pwdb.GormToErrcode(err)
		}
		activity := pwdb.Activity{
			Kind:                 pwdb.Activity_OrganizationInviteSend,
			AuthorID:             userID,
			UserID:               inviteUserID,
			OrganizationID:       organizationID,
			OrganizationMemberID: organizationOwner.OrganizationID,
		}
		return tx.Create(&activity).Error
	})
	if err != nil {
		return nil, pwdb.GormToErrcode(err)
	}

	// FIXME: Notify invited user

	ret := OrganizationSendInvite_Output{
		OrganizationInvite: &organizationInvite,
	}

	return &ret, nil
}

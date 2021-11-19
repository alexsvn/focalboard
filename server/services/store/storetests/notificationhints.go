// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package storetests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/focalboard/server/model"
	"github.com/mattermost/focalboard/server/services/store"
	"github.com/mattermost/focalboard/server/utils"
)

func StoreTestNotificationHintsStore(t *testing.T, setup func(t *testing.T) (store.Store, func())) {
	container := store.Container{
		WorkspaceID: "0",
	}

	t.Run("UpsertNotificationHint", func(t *testing.T) {
		store, tearDown := setup(t)
		defer tearDown()
		testUpsertNotificationHint(t, store, container)
	})

	t.Run("DeleteNotificationHint", func(t *testing.T) {
		store, tearDown := setup(t)
		defer tearDown()
		testDeleteNotificationHint(t, store, container)
	})

	t.Run("GetNotificationHint", func(t *testing.T) {
		store, tearDown := setup(t)
		defer tearDown()
		testGetNotificationHint(t, store, container)
	})

	t.Run("GetNextNotificationHint", func(t *testing.T) {
		store, tearDown := setup(t)
		defer tearDown()
		testGetNextNotificationHint(t, store, container)
	})
}

func testUpsertNotificationHint(t *testing.T, store store.Store, container store.Container) {
	t.Run("create notification hint", func(t *testing.T) {
		hint := &model.NotificationHint{
			BlockType:    model.TypeCard,
			BlockID:      utils.NewID(utils.IDTypeBlock),
			ModifiedByID: utils.NewID(utils.IDTypeUser),
			WorkspaceID:  container.WorkspaceID,
		}

		hintNew, err := store.UpsertNotificationHint(hint, time.Second*15)
		require.NoError(t, err, "upsert notification hint should not error")
		assert.Equal(t, hint.BlockID, hintNew.BlockID)
		assert.NoError(t, hintNew.IsValid())
	})

	t.Run("duplicate notification hint", func(t *testing.T) {
		hint := &model.NotificationHint{
			BlockType:    model.TypeCard,
			BlockID:      utils.NewID(utils.IDTypeBlock),
			ModifiedByID: utils.NewID(utils.IDTypeUser),
			WorkspaceID:  container.WorkspaceID,
		}
		hintNew, err := store.UpsertNotificationHint(hint, time.Second*15)
		require.NoError(t, err, "upsert notification hint should not error")

		// sleep a short time so the notify_at timestamps won't collide
		time.Sleep(time.Millisecond * 20)

		hint = &model.NotificationHint{
			BlockType:    model.TypeCard,
			BlockID:      hintNew.BlockID,
			ModifiedByID: hintNew.ModifiedByID,
			WorkspaceID:  container.WorkspaceID,
		}
		hintDup, err := store.UpsertNotificationHint(hint, time.Second*15)

		require.NoError(t, err, "upsert notification hint should not error")
		// the create_at fields should match, but notify_at should be updated
		assert.Equal(t, hintNew.CreateAt, hintDup.CreateAt)
		assert.Greater(t, hintDup.NotifyAt, hintNew.NotifyAt)
	})

	t.Run("invalid notification hint", func(t *testing.T) {
		hint := &model.NotificationHint{}

		_, err := store.UpsertNotificationHint(hint, time.Second*15)
		assert.ErrorAs(t, err, &model.ErrInvalidNotificationHint{}, "invalid notification hint should error")

		hint.BlockType = model.TypeBoard
		_, err = store.UpsertNotificationHint(hint, time.Second*15)
		assert.ErrorAs(t, err, &model.ErrInvalidNotificationHint{}, "invalid notification hint should error")

		hint.WorkspaceID = container.WorkspaceID
		_, err = store.UpsertNotificationHint(hint, time.Second*15)
		assert.ErrorAs(t, err, &model.ErrInvalidNotificationHint{}, "invalid notification hint should error")

		hint.ModifiedByID = utils.NewID(utils.IDTypeUser)
		_, err = store.UpsertNotificationHint(hint, time.Second*15)
		assert.ErrorAs(t, err, &model.ErrInvalidNotificationHint{}, "invalid notification hint should error")

		hint.BlockID = utils.NewID(utils.IDTypeBlock)
		hintNew, err := store.UpsertNotificationHint(hint, time.Second*15)
		assert.NoError(t, err, "valid notification hint should not error")
		assert.NoError(t, hintNew.IsValid(), "created notification hint should be valid")
	})
}

func testDeleteNotificationHint(t *testing.T, store store.Store, container store.Container) {
	t.Run("delete notification hint", func(t *testing.T) {
		hint := &model.NotificationHint{
			BlockType:    model.TypeCard,
			BlockID:      utils.NewID(utils.IDTypeBlock),
			ModifiedByID: utils.NewID(utils.IDTypeUser),
			WorkspaceID:  container.WorkspaceID,
		}
		hintNew, err := store.UpsertNotificationHint(hint, time.Second*15)
		require.NoError(t, err, "create notification hint should not error")

		// check the notification hint exists
		hint, err = store.GetNotificationHint(container, hintNew.BlockID)
		require.NoError(t, err, "get notification hint should not error")
		assert.Equal(t, hintNew.BlockID, hint.BlockID)
		assert.Equal(t, hintNew.CreateAt, hint.CreateAt)

		err = store.DeleteNotificationHint(container, hintNew.BlockID)
		require.NoError(t, err, "delete notification hint should not error")

		// check the notification hint was deleted
		hint, err = store.GetNotificationHint(container, hintNew.BlockID)
		require.True(t, store.IsErrNotFound(err), "error should be of type store.ErrNotFound")
		assert.Nil(t, hint)
	})

	t.Run("delete non-existent notification hint", func(t *testing.T) {
		err := store.DeleteNotificationHint(container, "bogus")
		require.True(t, store.IsErrNotFound(err), "error should be of type store.ErrNotFound")
	})
}

func testGetNotificationHint(t *testing.T, store store.Store, container store.Container) {
	t.Run("get notification hint", func(t *testing.T) {
		hint := &model.NotificationHint{
			BlockType:    model.TypeCard,
			BlockID:      utils.NewID(utils.IDTypeBlock),
			ModifiedByID: utils.NewID(utils.IDTypeUser),
			WorkspaceID:  container.WorkspaceID,
		}
		hintNew, err := store.UpsertNotificationHint(hint, time.Second*15)
		require.NoError(t, err, "create notification hint should not error")

		// make sure notification hint can be fetched
		hint, err = store.GetNotificationHint(container, hintNew.BlockID)
		require.NoError(t, err, "get notification hint should not error")
		assert.Equal(t, hintNew, hint)
	})

	t.Run("get non-existent notification hint", func(t *testing.T) {
		hint, err := store.GetNotificationHint(container, "bogus")
		require.True(t, store.IsErrNotFound(err), "error should be of type store.ErrNotFound")
		assert.Nil(t, hint, "hint should be nil")
	})
}

func testGetNextNotificationHint(t *testing.T, store store.Store, container store.Container) {
	t.Run("get next notification hint", func(t *testing.T) {
		const loops = 5
		ids := [5]string{}
		modifiedBy := utils.NewID(utils.IDTypeUser)

		// create some hints with unique notifyAt
		for i := 0; i < loops; i++ {
			hint := &model.NotificationHint{
				BlockType:    model.TypeCard,
				BlockID:      utils.NewID(utils.IDTypeBlock),
				ModifiedByID: modifiedBy,
				WorkspaceID:  container.WorkspaceID,
			}
			hintNew, err := store.UpsertNotificationHint(hint, time.Second*15)
			require.NoError(t, err, "create notification hint should not error")

			ids[i] = hintNew.BlockID
			time.Sleep(time.Millisecond * 20) // ensure next timestamp is unique
		}

		// check the hints come back in the right order
		notifyAt := utils.GetMillisForTime(time.Now().Add(time.Millisecond * 50))
		for i := 0; i < loops; i++ {
			hint, err := store.GetNextNotificationHint(false)
			require.NoError(t, err, "get next notification hint should not error")
			require.NotNil(t, hint, "get next notification hint should not return nil")
			assert.Equal(t, ids[i], hint.BlockID)
			assert.Less(t, notifyAt, hint.NotifyAt)
			notifyAt = hint.NotifyAt

			err = store.DeleteNotificationHint(container, hint.BlockID)
			require.NoError(t, err, "delete notification hint should not error")
		}
	})

	t.Run("get next notification hint from empty table", func(t *testing.T) {
		// empty the table
		err := emptyNotificationHintTable(store)
		require.NoError(t, err, "emptying notification hint table should not error")

		for {
			hint, err2 := store.GetNextNotificationHint(false)
			if store.IsErrNotFound(err2) {
				break
			}
			require.NoError(t, err2, "get next notification hint should not error")

			err2 = store.DeleteNotificationHint(container, hint.BlockID)
			require.NoError(t, err2, "delete notification hint should not error")
		}

		_, err = store.GetNextNotificationHint(false)
		require.True(t, store.IsErrNotFound(err), "error should be of type store.ErrNotFound")
	})

	t.Run("get next notification hint and remove", func(t *testing.T) {
		// empty the table
		err := emptyNotificationHintTable(store)
		require.NoError(t, err, "emptying notification hint table should not error")

		hint := &model.NotificationHint{
			BlockType:    model.TypeCard,
			BlockID:      utils.NewID(utils.IDTypeBlock),
			ModifiedByID: utils.NewID(utils.IDTypeUser),
			WorkspaceID:  container.WorkspaceID,
		}
		hintNew, err := store.UpsertNotificationHint(hint, time.Second*1)
		require.NoError(t, err, "create notification hint should not error")

		hintDeleted, err := store.GetNextNotificationHint(true)
		require.NoError(t, err, "get next notification hint should not error")
		require.NotNil(t, hintDeleted, "get next notification hint should not return nil")
		assert.Equal(t, hintNew.BlockID, hintDeleted.BlockID)

		// should be no hint left
		_, err = store.GetNextNotificationHint(false)
		require.True(t, store.IsErrNotFound(err), "error should be of type store.ErrNotFound")
	})
}

func emptyNotificationHintTable(store store.Store) error {
	for {
		hint, err := store.GetNextNotificationHint(false)
		if store.IsErrNotFound(err) {
			break
		}

		if err != nil {
			return err
		}

		c := containerForWorkspace(hint.WorkspaceID)

		err = store.DeleteNotificationHint(c, hint.BlockID)
		if err != nil {
			return err
		}
	}
	return nil
}

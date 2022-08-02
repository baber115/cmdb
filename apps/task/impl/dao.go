package impl

import (
	"context"

	"codeup.aliyun.com/baber/go/cmdb/apps/task"
)

func (i *impl) insert(ctx context.Context, t *task.Task) error {
	stmt, err := i.db.Prepare(insertTaskSQL)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(
		t.Id, t.Data.Region, t.Data.ResourceType, t.Data.SecretId, t.SecretDescription, t.Data.Timeout,
		t.Status.Stage, t.Status.Message, t.Status.StartAt, t.Status.EndAt, t.Status.TotalSucceed, t.Status.TotalFailed,
	)
	if err != nil {
		return err
	}

	return nil
}

func (i *impl) update(ctx context.Context, t *task.Task) error {
	stmt, err := i.db.Prepare(updateTaskSQL)
	if err != nil {
		return err
	}
	defer stmt.Close()
	status := t.Status
	_, err = stmt.Exec(
		status.Stage, status.Message, status.EndAt, status.TotalSucceed, status.TotalFailed, t.Id,
	)
	if err != nil {
		return err
	}

	return nil
}

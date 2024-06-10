package query

import (
	"kdmid-queue-checker/domain/notification"
	"reflect"
	"testing"
)

func TestListUsers_mergeAndCastUsers(t *testing.T) {
	t.Parallel()

	type args struct {
		activeRecipients []notification.Recipient
		usersWithCrawls  []int64
	}
	tests := []struct {
		name string
		args args
		want []User
	}{
		{
			name: "Basic Functionality",
			args: args{
				activeRecipients: []notification.Recipient{{TelegramID: 1}, {TelegramID: 2}},
				usersWithCrawls:  []int64{2, 3},
			},
			want: []User{
				{TelegramID: 1, Active: true, HasCrawls: false},
				{TelegramID: 2, Active: true, HasCrawls: true},
				{TelegramID: 3, Active: false, HasCrawls: true},
			},
		},
		{
			name: "No Overlap",
			args: args{
				activeRecipients: []notification.Recipient{{TelegramID: 1}},
				usersWithCrawls:  []int64{2},
			},
			want: []User{
				{TelegramID: 1, Active: true, HasCrawls: false},
				{TelegramID: 2, Active: false, HasCrawls: true},
			},
		},
		{
			name: "Complete Overlap",
			args: args{
				activeRecipients: []notification.Recipient{{TelegramID: 1}},
				usersWithCrawls:  []int64{1},
			},
			want: []User{
				{TelegramID: 1, Active: true, HasCrawls: true},
			},
		},
		{
			name: "Empty Inputs",
			args: args{
				activeRecipients: []notification.Recipient{},
				usersWithCrawls:  []int64{},
			},
			want: []User{},
		},
		{
			name: "Single Element Lists",
			args: args{
				activeRecipients: []notification.Recipient{{TelegramID: 1}},
				usersWithCrawls:  []int64{1},
			},
			want: []User{
				{TelegramID: 1, Active: true, HasCrawls: true},
			},
		},
		{
			name: "Duplicate Entries in Active Recipients",
			args: args{
				activeRecipients: []notification.Recipient{{TelegramID: 1}, {TelegramID: 1}},
				usersWithCrawls:  []int64{2},
			},
			want: []User{
				{TelegramID: 1, Active: true, HasCrawls: false},
				{TelegramID: 2, Active: false, HasCrawls: true},
			},
		},
		{
			name: "Duplicate Entries in Users With Crawls",
			args: args{
				activeRecipients: []notification.Recipient{{TelegramID: 1}},
				usersWithCrawls:  []int64{2, 2},
			},
			want: []User{
				{TelegramID: 1, Active: true, HasCrawls: false},
				{TelegramID: 2, Active: false, HasCrawls: true},
			},
		},
		{
			name: "Non-Sequential IDs",
			args: args{
				activeRecipients: []notification.Recipient{{TelegramID: 1}, {TelegramID: 1000}},
				usersWithCrawls:  []int64{500, 1000},
			},
			want: []User{
				{TelegramID: 1, Active: true, HasCrawls: false},
				{TelegramID: 500, Active: false, HasCrawls: true},
				{TelegramID: 1000, Active: true, HasCrawls: true},
			},
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := &ListUsersHandler{}

			if got := h.mergeAndCastUsers(tt.args.activeRecipients, tt.args.usersWithCrawls); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeAndCastUsers() = %v, want %v", got, tt.want)
			}
		})
	}
}

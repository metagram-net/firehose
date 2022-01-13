create type drop_status as enum ('unread', 'read', 'saved');

alter table drops alter column status type drop_status using status::drop_status;

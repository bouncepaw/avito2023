create table segments
(
	id                serial primary key,
	name              text unique,
	deleted           boolean
		default false,
	automatic_percent smallint
		default 0
		check ( automatic_percent >= 0 and automatic_percent <= 100 )
);

create table users_to_segments
(
	user_id    integer,
	segment_id integer
		references segments (id),
	unique (user_id, segment_id)
);

create table delayed_removals
(
	stamp      timestamp,
	user_id    integer,
	segment_id integer
		references segments (id)
);

create type operation_type as enum ( 'add', 'remove' );

create table operation_history
(
	stamp      timestamp with time zone default now(),
	user_id    integer,
	segment_id integer
		references segments (id),
	operation  operation_type
);

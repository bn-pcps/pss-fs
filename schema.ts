import {
	pgTable,
	uuid,
	varchar,
	integer,
	timestamp,
	text,
	boolean,
	bigint,
	index,
	uniqueIndex,
	check
} from 'drizzle-orm/pg-core';
import { sql } from 'drizzle-orm';

// ps_plans
export const ps_plans = pgTable(
	'ps_plans',
	{
		id: integer('id').primaryKey(),
		polar_id: varchar('polar_id', { length: 255 }),
		plan_name: varchar('plan_name', { length: 100 }).notNull(),
		quota: bigint('quota', { mode: 'number' }).notNull() // in MB
	},
	(table) => [
		index('ps_plans_polar_id_idx').on(table.polar_id),
		index('ps_plans_plan_name_idx').on(table.plan_name),
		check('quota_positive', sql`${table.quota} >= 0`)
	]
);

// ps_users
export const ps_users = pgTable(
	'ps_users',
	{
		id: uuid('id').primaryKey().defaultRandom(),
		google_id: varchar('google_id', { length: 255 }).unique().notNull(),
		name: varchar('name', { length: 255 }).notNull(),
		email: varchar('email', { length: 255 }).unique().notNull(),
		avatar_url: varchar('avatar_url', { length: 255 }),
		created_at: timestamp('created_at', { withTimezone: true }).defaultNow().notNull(),
		updated_at: timestamp('updated_at', { withTimezone: true }).defaultNow().notNull()
	},
	(table) => [
		index('ps_users_email_idx').on(table.email),
		index('ps_users_google_id_idx').on(table.google_id),
		index('ps_users_created_at_idx').on(table.created_at)
	]
);

// ps_user_plan
export const ps_user_plan = pgTable('ps_user_plan', {
	user_id: uuid('user_id')
		.references(() => ps_users.id, { onDelete: 'cascade' })
		.primaryKey(),
	plan_id: integer('plan_id')
		.references(() => ps_plans.id, { onDelete: 'restrict' })
		.default(1)
		.notNull(),
	created_at: timestamp('created_at', { withTimezone: true }).defaultNow().notNull(),
	updated_at: timestamp('updated_at', { withTimezone: true }).defaultNow().notNull(),
	expires_at: timestamp('expires_at', { withTimezone: true }),
	subscription_id: varchar('subscription_id', { length: 255 })
});

// ps_used_quota
export const ps_used_quota = pgTable(
	'ps_used_quota',
	{
		user_id: uuid('user_id')
			.references(() => ps_users.id, { onDelete: 'cascade' })
			.primaryKey(),
		used_quota: bigint('used_quota', { mode: 'number' }).default(0).notNull(), // in MB
		last_updated: timestamp('last_updated', { withTimezone: true }).defaultNow().notNull()
	},
	(table) => [check('used_quota_positive', sql`${table.used_quota} >= 0`)]
);

// ps_shares
export const ps_shares = pgTable(
	'ps_shares',
	{
		id: uuid('id').primaryKey().defaultRandom(),
		user_id: uuid('user_id')
			.references(() => ps_users.id, { onDelete: 'cascade' })
			.notNull(),
		created_at: timestamp('created_at', { withTimezone: true }).defaultNow().notNull(),
		updated_at: timestamp('updated_at', { withTimezone: true }).defaultNow().notNull(),
		deleted_at: timestamp('deleted_at', { withTimezone: true }),
		title: varchar('title', { length: 255 }).notNull(),
		description: text('description'),
		file_count: integer('file_count').default(0).notNull(),
		size: bigint('size', { mode: 'number' }).default(0).notNull(),
		download_count: integer('download_count').default(0).notNull(),
		view_count: integer('view_count').default(0).notNull(),
		is_public: boolean('is_public').default(true).notNull()
	},
	(table) => [
		index('ps_shares_user_id_idx').on(table.user_id),
		index('ps_shares_created_at_idx').on(table.created_at),
		index('ps_shares_deleted_at_idx').on(table.deleted_at),
		index('ps_shares_is_public_idx').on(table.is_public),
		check('file_size_positive', sql`${table.size} >= 0`),
		check('file_count_positive', sql`${table.file_count} >= 0`),
		check('download_count_positive', sql`${table.download_count} >= 0`),
		check('view_count_positive', sql`${table.view_count} >= 0`)
	]
);

// ps_files
export const ps_files = pgTable(
	'ps_files',
	{
		id: uuid('id').primaryKey().defaultRandom(),
		created_at: timestamp('created_at', { withTimezone: true }).defaultNow().notNull(),
		deleted_at: timestamp('deleted_at', { withTimezone: true }),
		s3_key: varchar('s3_key', { length: 512 }), // will only be used for s3 storage
		share_id: uuid('share_id')
			.references(() => ps_shares.id, { onDelete: 'cascade' })
			.notNull(),
		file_name: varchar('file_name', { length: 255 }).notNull(),
		mimetype: varchar('mimetype', { length: 100 }).notNull(),
		hash: varchar('hash', { length: 255 }).notNull(),
		size: bigint('size', { mode: 'number' }).notNull()
	},
	(table) => [
		index('ps_files_share_id_idx').on(table.share_id),
		index('ps_files_hash_idx').on(table.hash),
		index('ps_files_deleted_at_idx').on(table.deleted_at),
		check('file_size_positive', sql`${table.size} >= 0`)
	]
);

// ps_share_settings
export const ps_share_settings = pgTable(
	'ps_share_settings',
	{
		id: uuid('id').primaryKey().defaultRandom(),
		share_id: uuid('share_id')
			.references(() => ps_shares.id, { onDelete: 'cascade' })
			.unique()
			.notNull(),
		expiry: timestamp('expiry', { withTimezone: true }),
		password_hash: varchar('password_hash', { length: 255 }),
		download_limit: integer('download_limit'),
		custom_slug: varchar('custom_slug', { length: 255 }).unique(),
		created_at: timestamp('created_at', { withTimezone: true }).defaultNow().notNull(),
		updated_at: timestamp('updated_at', { withTimezone: true }).defaultNow().notNull()
	},
	(table) => [
		index('ps_share_settings_share_id_idx').on(table.share_id),
		uniqueIndex('ps_share_settings_custom_slug_idx').on(table.custom_slug),
		index('ps_share_settings_expiry_idx').on(table.expiry),
		check(
			'download_limit_positive',
			sql`${table.download_limit} IS NULL OR ${table.download_limit} > 0`
		)
	]
);

// ps_upload_signatures
export const ps_upload_signatures = pgTable(
	'ps_upload_signatures',
	{
		id: uuid('id').primaryKey().defaultRandom(),
		share_id: uuid('share_id')
			.references(() => ps_shares.id, { onDelete: 'cascade' })
			.notNull(),
		signature: varchar('signature', { length: 512 }).notNull().unique(),
		expiry: timestamp('expiry', { withTimezone: true }).notNull(),
		is_used: boolean('is_used').default(false).notNull(), // deprecated, will be removed
		expected_file_count: integer('expected_file_count').notNull(),
		expected_file_size: bigint('expected_file_size', { mode: 'number' }).notNull(), // in MB, rounded up to nearest MB
		used_at: timestamp('used_at', { withTimezone: true }),
		created_at: timestamp('created_at', { withTimezone: true }).defaultNow().notNull()
	},
	(table) => [
		index('ps_upload_signatures_share_id_idx').on(table.share_id),
		index('ps_upload_signatures_expiry_idx').on(table.expiry),
		index('ps_upload_signatures_is_used_idx').on(table.is_used),
		uniqueIndex('ps_upload_signatures_signature_idx').on(table.signature)
	]
);

// ps_download_signatures
export const ps_download_signatures = pgTable(
	'ps_download_signatures',
	{
		id: uuid('id').primaryKey().defaultRandom(),
		share_id: uuid('share_id')
			.references(() => ps_shares.id, { onDelete: 'cascade' })
			.notNull(),
		signature: varchar('signature', { length: 512 }).notNull().unique(),
		expiry: timestamp('expiry', { withTimezone: true }).notNull(),
		is_used: boolean('is_used').default(false).notNull(),
		used_at: timestamp('used_at', { withTimezone: true }),
		created_at: timestamp('created_at', { withTimezone: true }).defaultNow().notNull()
	},
	(table) => [
		index('ps_download_signatures_share_id_idx').on(table.share_id),
		index('ps_download_signatures_expiry_idx').on(table.expiry),
		index('ps_download_signatures_is_used_idx').on(table.is_used),
		uniqueIndex('ps_download_signatures_signature_idx').on(table.signature)
	]
);

// ps_download_analytics
export const ps_download_analytics = pgTable(
	'ps_download_analytics',
	{
		id: uuid('id').primaryKey().defaultRandom(),
		share_id: uuid('share_id')
			.references(() => ps_shares.id, { onDelete: 'cascade' })
			.notNull(),
		file_id: uuid('file_id').references(() => ps_files.id, {
			onDelete: 'set null'
		}),
		timestamp: timestamp('timestamp', { withTimezone: true }).defaultNow().notNull(),
		ip_address: varchar('ip_address', { length: 45 }),
		user_agent: varchar('user_agent', { length: 512 }),
		country: varchar('country', { length: 2 }), // ISO country code (but how to get this?)
		city: varchar('city', { length: 100 })
	},
	(table) => [
		index('ps_download_analytics_share_id_idx').on(table.share_id),
		index('ps_download_analytics_file_id_idx').on(table.file_id),
		index('ps_download_analytics_timestamp_idx').on(table.timestamp),
		index('ps_download_analytics_ip_address_idx').on(table.ip_address)
	]
);

// ps_visit_analytics
export const ps_visit_analytics = pgTable(
	'ps_visit_analytics',
	{
		id: uuid('id').primaryKey().defaultRandom(),
		share_id: uuid('share_id')
			.references(() => ps_shares.id, { onDelete: 'cascade' })
			.notNull(),
		timestamp: timestamp('timestamp', { withTimezone: true }).defaultNow().notNull(),
		ip_address: varchar('ip_address', { length: 45 }),
		user_agent: varchar('user_agent', { length: 512 }),
		referrer: varchar('referrer', { length: 512 }),
		country: varchar('country', { length: 2 }),
		city: varchar('city', { length: 100 })
	},
	(table) => [
		index('ps_visit_analytics_share_id_idx').on(table.share_id),
		index('ps_visit_analytics_timestamp_idx').on(table.timestamp),
		index('ps_visit_analytics_ip_address_idx').on(table.ip_address)
	]
);

// Export all tables for easy import
export const tables = {
	ps_plans,
	ps_users,
	ps_user_plan,
	ps_used_quota,
	ps_shares,
	ps_files,
	ps_share_settings,
	ps_upload_signatures,
	ps_download_signatures,
	ps_download_analytics,
	ps_visit_analytics
};

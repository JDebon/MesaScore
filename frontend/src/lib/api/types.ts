export interface User {
	id: string;
	username: string;
	display_name: string;
	avatar_url: string | null;
}

export interface UserProfile extends User {
	created_at: string;
}

export interface Party {
	id: string;
	name: string;
	description: string | null;
	admin: User;
	invite_code: string;
	member_count: number;
	created_at: string;
}

export interface LoginResponse {
	token: string;
	user: User;
}

export interface DashboardResponse {
	parties: { id: string; name: string; member_count: number; last_session_at: string | null }[];
	pending_invites: PendingInvite[];
	global_stats: { total_sessions: number; total_wins: number; current_streak: number };
}

export interface PendingInvite {
	id: string;
	party: { id: string; name: string };
	invited_by: { id: string; display_name: string };
	created_at: string;
}

export interface SessionSummary {
	id: string;
	game: { id: string; name: string; cover_image_url: string | null };
	session_type: string;
	played_at: string;
	duration_minutes: number | null;
	winners: { id: string; display_name: string; avatar_url?: string | null }[];
	participant_count: number;
}

export interface SessionDetail {
	id: string;
	game: { id: string; name: string; cover_image_url: string | null };
	session_type: 'competitive' | 'team' | 'cooperative' | 'score';
	played_at: string;
	duration_minutes: number | null;
	notes: string | null;
	brought_by: { id: string; display_name: string } | null;
	created_by: { id: string; display_name: string };
	created_at: string;
	participants: SessionParticipant[];
}

export interface SessionParticipant {
	user: User & { avatar_url: string | null };
	team_name: string | null;
	rank: number | null;
	score: number | null;
	result: 'win' | 'loss' | 'draw' | null;
}

export interface Game {
	id: string;
	bgg_id: number | null;
	name: string;
	cover_image_url: string | null;
	min_players: number | null;
	max_players: number | null;
	bgg_rating: number | null;
	session_count: number;
	in_my_collection: boolean;
}

export interface GameDetail extends Game {
	description: string | null;
	bgg_fetched_at: string | null;
	added_by: { id: string; display_name: string };
	created_at: string;
	owners: { id: string; display_name: string; avatar_url: string | null }[];
}

export interface LeaderboardEntry {
	user: User & { avatar_url: string | null };
	wins: number;
	sessions: number;
	win_rate: number;
}

export interface UserStats {
	user: User & { avatar_url: string | null };
	total_sessions: number;
	total_wins: number;
	win_rate: number;
	current_streak: number;
	best_streak: number;
	most_played_game: { id: string; name: string; session_count: number } | null;
	best_win_rate_game: { id: string; name: string; win_rate: number } | null;
	nemesis: { id: string; display_name: string; losses_against: number } | null;
	punching_bag: { id: string; display_name: string; wins_against: number } | null;
	per_game: PerGameStat[];
	head_to_head: HeadToHeadEntry[];
}

export interface PerGameStat {
	game: { id: string; name: string; cover_image_url: string | null };
	sessions: number;
	wins: number;
	win_rate: number;
}

export interface HeadToHeadEntry {
	opponent: User & { avatar_url: string | null };
	sessions_together: number;
	this_user_wins: number;
	opponent_wins: number;
}

export interface CollectionItem {
	game_id: string;
	name: string;
	cover_image_url: string | null;
	added_at: string;
}

export interface PaginatedResponse<T> {
	data: T[];
	total: number;
	page: number;
	per_page: number;
}

export interface PartyDashboard {
	total_sessions: number;
	total_unique_games: number;
	total_members: number;
	current_leader: { user: User & { avatar_url: string | null }; wins: number } | null;
	most_played_game: {
		id: string;
		name: string;
		cover_image_url: string | null;
		session_count: number;
	} | null;
	sessions_per_month: { month: string; count: number }[];
	recent_sessions: {
		id: string;
		game: { id: string; name: string; cover_image_url: string | null };
		played_at: string;
		session_type: string;
		winners: { id: string; display_name: string }[];
	}[];
}

export interface PartyMember {
	id: string;
	username: string;
	display_name: string;
	avatar_url: string | null;
	is_admin: boolean;
	joined_at: string;
}

export interface PartyInvite {
	id: string;
	invited_user: { id: string; username: string; display_name: string };
	status: 'pending' | 'accepted' | 'declined';
	created_at: string;
}

export interface MembersResponse {
	members: PartyMember[];
	invites: PartyInvite[];
}

export interface BggSearchResult {
	bgg_id: number;
	name: string;
	year: number | null;
	thumbnail_url: string | null;
}

export interface BGGGameDetail {
	bgg_id: number;
	name: string;
	year: number | null;
	description: string | null;
	cover_image_url: string | null;
	min_players: number | null;
	max_players: number | null;
	bgg_rating: number | null;
}

export interface AvailableGame {
	id: string;
	name: string;
	cover_image_url: string | null;
	owners: { id: string; display_name: string }[];
}

export interface ParticipantInput {
	user_id: string;
	team_name: string | null;
	rank: number | null;
	score: number | null;
	result: 'win' | 'loss' | 'draw' | null;
}

export interface GamePartyStats {
	game: { id: string; name: string };
	total_sessions: number;
	last_played_at: string | null;
	sessions_per_month: { month: string; count: number }[];
	leaderboard: LeaderboardEntry[];
}

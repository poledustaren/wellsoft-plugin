// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

export interface MattermostPost {
    id: string;
    message: string;
    user_id: string;
    channel_id: string;
    create_at: number;
    update_at: number;
    file_ids: string[];
    type: string;
}

export interface MattermostUser {
    id: string;
    username: string;
    email: string;
    nickname: string;
    first_name: string;
    last_name: string;
    position: string;
    roles: string;
    locale: string;
    timezone: object;
    is_bot: boolean;
    delete_at: number;
}

export interface MattermostChannel {
    id: string;
    create_at: number;
    update_at: number;
    delete_at: number;
    team_id: string;
    type: string;
    display_name: string;
    name: string;
    header: string;
    purpose: string;
    last_post_at: number;
    total_msg_count: number;
    extra_update_at: number;
    creator_id: string;
}

export interface ForwardRequest {
    post_id: string;
    recipient: string;
    sender_id: string;
}

export {};

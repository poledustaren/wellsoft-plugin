// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React from 'react';
import {useSelector} from 'react-redux';

import {TelegramMessageBlock} from './TelegramMessageBlock';

import {useMessages} from '../hooks/useMessages';
import {formatTelegramText} from '../utils/telegram_formatter';

interface Message {
    id: string;
    content: string;
    userId: string;
    timestamp: string;
}

export const ChatView: React.FC = () => {
    const messages = useMessages(); // ваш хук для получения сообщений
    const currentUserId = useSelector((state: any) => state.entities.users.currentUserId);
    return (
        <div className='telegram-chat-container'>
            {messages.map(((msg: Message, index: number) => (
                <div
                    key={msg.id}
                    className='telegram-message-group'
                >
                    <TelegramMessageBlock
                        message={formatTelegramText(msg.content)}
                        isOutgoing={msg.userId === currentUserId}
                        timestamp={msg.timestamp}
                        userId={msg.userId}
                    />
                    {index < messages.length - 1 && (
                        <div className='telegram-message-separator'/>
                    )}
                </div>
            )))}
        </div>
    );
};

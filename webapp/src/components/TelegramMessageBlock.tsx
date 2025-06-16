// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React from 'react';
import '../styles/message_block.css';

interface MessageBlockProps {
    message: string;
    isOutgoing: boolean;
    timestamp: string;
    userId: string;
}

export const TelegramMessageBlock: React.FC<MessageBlockProps> = ({
    message,
    isOutgoing,
    timestamp,
    userId,
}) => {
    return (
        <div className={`telegram-message-block ${isOutgoing ? 'outgoing' : 'incoming'}`}>
            <div className='message-content'>
                {message}
            </div>
            <div className='message-meta'>
                <span className='timestamp'>{timestamp}</span>
            </div>
        </div>
    );
};

// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import Markdown from 'markdown-to-jsx';
import React from 'react';

type Props = {
    post: any;
};

export const ForwardedMessage: React.FC<Props> = ({post}) => {
    const props = post.props || {};
    const displayName = props.original_display_name || 'Неизвестно';
    const username = props.original_username || '';
    const createAt = props.original_create_at ? new Date(props.original_create_at).toLocaleString() : '';
    const message = post.message || '';
    const fileIds = post.file_ids || [];

    return (
        <div className='forwarded-message'>
            <span className='forwarded-author'>{username}</span>
            {createAt && <span className='forwarded-time'>{createAt}</span>}
            <div className='forwarded-text'>
                <Markdown>{message}</Markdown>
            </div>
            {fileIds.length > 0 && (
                <div className='forwarded-files'>
                    <span style={{fontSize: 13, color: '#888'}}>Вложения:</span>
                    {fileIds.map((fileId: string, idx: number) => (
                        <a
                            key={fileId}
                            href={`/api/v4/files/${fileId}`}
                            target='_blank'
                            rel='noopener noreferrer'
                            className='forwarded-file-link'
                        >
                            📎 Файл {idx + 1}
                        </a>
                    ))}
                </div>
            )}
        </div>
    );
};

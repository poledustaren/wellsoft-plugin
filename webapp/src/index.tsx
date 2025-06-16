// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React from 'react';
import {render, unmountComponentAtNode} from 'react-dom';

import './styles.css';
import './styles/dark_theme.css';
import {ChatView} from './components/ChatView';
import {ForwardedMessage} from './components/ForwardedMessage';
import {ForwardModal} from './components/ForwardModal';
import manifest from './manifest';

class ForwardPlugin {
    root: HTMLElement | null = null;

    initialize(registry: any, store: any) {
        registry.registerPostDropdownMenuAction(
            'Переслать сообщение',
            (postId: string) => {
                this.showForwardModal(postId);
            },
            () => true,
        );

        // регистрация компонентов и стилей
        registry.registerNeedsTeamRoute(
            '/telegram-chat',
            ChatView,
        );

        registry.registerPostTypeComponent('custom_forwarded', ForwardedMessage);
    }

    showForwardModal(postId: string) {
        if (this.root) {
            document.body.removeChild(this.root);
        }
        this.root = document.createElement('div');
        document.body.appendChild(this.root);

        const handleClose = () => {
            if (this.root) {
                unmountComponentAtNode(this.root);
                document.body.removeChild(this.root);
                this.root = null;
            }
        };

        const handleSubmit = async (recipient: string) => {
            await fetch(`/plugins/${manifest.id}/forward`, {
                method: 'POST',
                headers: {'Content-Type': 'application/json', 'X-Requested-With': 'XMLHttpRequest'},
                body: JSON.stringify({post_id: postId, recipient}),
            });
            handleClose();
        };

        render(
            <ForwardModal
                postId={postId}
                onSubmit={handleSubmit}
                onClose={handleClose}
            />,
            this.root,
        );
    }
}

declare global {
    interface Window {
        registerPlugin(id: string, plugin: ForwardPlugin): void;
    }
}

window.registerPlugin(manifest.id, new ForwardPlugin());
export {};

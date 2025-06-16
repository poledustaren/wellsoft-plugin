// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

export const formatTelegramText = (text: string): string => {
    return text.
        replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>'). // жирный
        replace(/__(.*?)__/g, '<u>$1</u>'). // подчеркнутый
        replace(/_(.*?)_/g, '<em>$1</em>'). // курсив
        replace(/~~(.*?)~~/g, '<del>$1</del>'). // зачеркнутый
        replace(/\|\|(.*?)\|\|/g, '<span class="spoiler">$1</span>'). // спойлер
        replace(/`(.*?)`/g, '<code>$1</code>'); // моноширинный
};

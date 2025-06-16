// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {useSelector} from 'react-redux';

export const useMessages = () => {
    return useSelector((state: any) => state.plugins.myPlugin.messages);
};

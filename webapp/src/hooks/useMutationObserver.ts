// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import type {MutableRefObject} from 'react';
import {useEffect} from 'react';

type MutationObserverOptions = {
    attributes?: boolean;
    characterData?: boolean;
    childList?: boolean;
    subtree?: boolean;
};

export function useMutationObserver(
    ref: MutableRefObject<HTMLElement | null>,
    callback: MutationCallback,
    options: MutationObserverOptions = {
        attributes: true,
        childList: true,
        subtree: true,
    },
) {
    useEffect(() => {
        if (!ref.current) {
            return;
        }
        console.log('use effects');
        const observer = new MutationObserver(callback);
        observer.observe(ref.current, options);

        return () => {
            observer.disconnect();
        };
    }, [ref, callback, options]);
}

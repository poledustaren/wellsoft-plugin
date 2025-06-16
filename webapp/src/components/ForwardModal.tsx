// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React, {useState, useEffect} from 'react';

import manifest from '../manifest';

type Option = {
    label: string;
    value: string;
};

type Props = {
    postId: string;
    onSubmit: (recipient: string) => void;
    onClose: () => void;
};

const API_URL = `/plugins/${manifest.id}/recipients_list`;

export const ForwardModal: React.FC<Props> = ({postId, onSubmit, onClose}) => {
    const [options, setOptions] = useState<Option[]>([]);
    const [search, setSearch] = useState('');
    const [selected, setSelected] = useState('');
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchOptions = async () => {
            setLoading(true);
            const resp = await fetch(API_URL);
            const data = await resp.json();
            const opts = data.map((item: any) => ({
                label: item.label,
                value: `${item.type}:${item.id}`,
            }));
            setOptions(opts);

            // Сразу выбираем первый элемент, если он есть
            if (opts.length > 0) {
                setSelected(opts[0].value);
            }
            setLoading(false);
        };
        fetchOptions();
    }, []);

    // Фильтрация на фронте
    const filteredOptions = options.filter((opt) =>
        opt.label.toLowerCase().includes(search.toLowerCase()),
    );

    // Если после фильтрации ничего не выбрано, выбираем первый из отфильтрованных
    useEffect(() => {
        if (filteredOptions.length > 0 && !filteredOptions.find((o) => o.value === selected)) {
            setSelected(filteredOptions[0].value);
        }
    }, [search, options]); // Обновлять при изменении поиска или списка

    return (
        <div className='forward-modal-backdrop'>
            <div
                className='forward-modal'
                style={{padding: 24, background: '#fff', borderRadius: 6, width: 400}}
            >
                <h3>Переслать сообщение</h3>
                <input
                    autoFocus={true}
                    type='text'
                    placeholder='Поиск пользователя или канала...'
                    value={search}
                    onChange={(e) => setSearch(e.target.value)}
                    style={{width: '100%', marginBottom: 12, padding: 8}}
                />
                {loading ? (
                    <div>Загрузка...</div>
                ) : (
                    <select
                        style={{width: '100%', padding: 8, marginBottom: 16}}
                        value={selected}
                        onChange={(e) => setSelected(e.target.value)}
                    >
                        {filteredOptions.length === 0 && (
                            <option value=''>Нет результатов</option>
                        )}
                        {filteredOptions.map((opt) => (
                            <option
                                key={opt.value}
                                value={opt.value}
                            >
                                {opt.label}
                            </option>
                        ))}
                    </select>
                )}
                <div style={{textAlign: 'right'}}>
                    <button
                        onClick={onClose}
                        style={{marginRight: 8}}
                    >Отмена</button>
                    <button
                        disabled={!selected}
                        onClick={() => onSubmit(selected)}
                    >
                        Переслать
                    </button>
                </div>
            </div>
        </div>
    );
};

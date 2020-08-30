import React from 'react';

const UserContext = React.createContext({
    id: undefined,
    name: undefined,
    token: undefined
});

export default UserContext;
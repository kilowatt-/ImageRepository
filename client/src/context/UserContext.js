import React, {useContext, useReducer} from 'react';

const reducer = (state, action) => {
    switch (action.type) {
        case 'setID':
            return {...state, id: action.id};
        case 'setName':
            return {...state, name: action.name}
        case 'setUser':
            return action.user;
        default:
            return state;
    }
}

const initialState = {
    id: undefined,
    name: undefined
};

const UserContext = React.createContext();

export const UserContextProvider = ({children}) => {
    const contextValue = useReducer(reducer, initialState);
    return (
        <UserContext.Provider value={contextValue}>
            {children}
        </UserContext.Provider>
    )
}

export const useUserContext = () => {
    return useContext(UserContext);
}
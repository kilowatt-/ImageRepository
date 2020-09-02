import Cookies from "universal-cookie";

const cookies = new Cookies();

export const getUserFromCookie = (dispatch) => {


    const cookieUser = cookies.get('userinfo');

    if (cookieUser) {
        const replaced = cookieUser.replace(/'/g, '"')


        const parsedCookieUser = JSON.parse(replaced);


        if (parsedCookieUser) {
            dispatch({type: 'setUser', user: parsedCookieUser});
        }
    }
}

export const resetCookie = (dispatch) => {
    cookies.remove('userinfo');
    cookies.remove('token');
    dispatch({type: 'reset'});
}
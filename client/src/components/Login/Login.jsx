import React, {useContext, useState} from 'react';
import Avatar from '@material-ui/core/Avatar';
import Button from '@material-ui/core/Button';
import CssBaseline from '@material-ui/core/CssBaseline';
import TextField from '@material-ui/core/TextField';
import Link from '@material-ui/core/Link';
import Grid from '@material-ui/core/Grid';
import LockOutlinedIcon from '@material-ui/icons/LockOutlined';
import Typography from '@material-ui/core/Typography';
import { makeStyles } from '@material-ui/core/styles';
import Container from '@material-ui/core/Container';
import {API_CONFIG} from "../../config/api";
import UserContext from "../../context/UserContext";
import CircularProgress from "@material-ui/core/CircularProgress";
import * as qs from "querystring";
import axios from 'axios';

const useStyles = makeStyles((theme) => ({
    paper: {
        marginTop: theme.spacing(8),
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
    },
    avatar: {
        margin: theme.spacing(1),
        backgroundColor: theme.palette.secondary.main,
    },
    form: {
        width: '100%', // Fix IE 11 issue.
        marginTop: theme.spacing(1),
    },
    submit: {
        margin: theme.spacing(3, 0, 2),
    },
    error: {
        color: 'red'
    }
}));

const Login = () => {
    const user = useContext(UserContext);

    const [loggingIn, setLoggingIn] = useState(false);
    const [errorMessage, setErrorMessage] = useState("");

    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");

    const [emailError, setEmailError] = useState(false);
    const [passwordError, setPasswordError] = useState(false);

    const handleEmailChange = (event) => {
        setEmail(event.target.value);
        if (event.target.value === '') {
          setEmailError(true);
        } else {
            setEmailError(false);
        }
    }

    const handlePasswordChange = (event) => {
        setPassword(event.target.value);

        if (event.target.value === '') {
            setPasswordError(true);
        } else {
            setPasswordError(false);
        }
    }


    const handleSubmit = async (event) => {
        event.preventDefault();

        if (!loggingIn) {
            setErrorMessage("");
            setLoggingIn(true);

            const config = {
                headers: {
                    'Content-Type': 'application/x-www-form-urlencoded',
                    "Access-Control-Allow-Origin": "*"
                }
            }

            const body = {
                email,
                password
            }

            try {
                const response = axios.post(`${API_CONFIG.base_url}/users/login`, qs.stringify(body), config);
            } catch (err) {
                console.log(err);
            }
        }

    }

    const classes = useStyles();

    return (
        <Container component="main" maxWidth="xs">
            <CssBaseline />
            <div className={classes.paper}>
                <Avatar className={classes.avatar}>
                    <LockOutlinedIcon />
                </Avatar>
                <Typography component="h1" variant="h5">
                    Sign in
                </Typography>
                <form className={classes.form} noValidate onSubmit={handleSubmit}>
                    <TextField
                        variant="outlined"
                        margin="normal"
                        required
                        fullWidth
                        id="email"
                        label="Email Address"
                        name="email"
                        autoComplete="email"
                        disabled={loggingIn}
                        value={email}
                        error={emailError}
                        onChange={handleEmailChange}
                        autoFocus
                    />
                    <TextField
                        variant="outlined"
                        margin="normal"
                        required
                        fullWidth
                        name="password"
                        label="Password"
                        type="password"
                        id="password"
                        error={passwordError}
                        disabled={loggingIn}
                        autoComplete="current-password"
                        value={password}
                        onChange={handlePasswordChange}
                    />
                    <Button
                        type="submit"
                        fullWidth
                        variant="contained"
                        color="primary"
                        className={classes.submit}
                        disabled={loggingIn}
                    >
                        {loggingIn ? <CircularProgress color={"secondary"}/> : "Sign In"}
                    </Button>
                    <Grid container>
                        <Grid item>
                            <Link href="/signup" variant="body2">
                                {"Don't have an account? Sign Up"}
                            </Link>
                        </Grid>
                    </Grid>
                    <p className={classes.error}>{errorMessage}</p>
                </form>
            </div>
        </Container>);
}

export default Login;
import {Button, Container, Typography} from "@material-ui/core";
import Header from "../Header/Header";
import React, {useState} from 'react';
import {useUserContext} from "../../context/UserContext";
import Grid from "@material-ui/core/Grid";
import UploadModal from "./UploadModal";

const Home = () => {
    const [user] = useUserContext();
    const [uploadModalVisible, setUploadModalVisible] = useState(false);

    return (
        <div>
            <Header />
            <Container>
                <Typography variant={"h4"}>Your feed</Typography>
                <Grid container>
                    <UploadModal open={uploadModalVisible} handleClose={() => setUploadModalVisible(false)} />
                    {user.name ?  (<Button variant={"contained"} onClick={() => setUploadModalVisible(true)}>+ Upload New Image</Button>) : null}
                    <Grid container spacing={2} justify={"center"}>
                        <Grid item></Grid>
                        <Grid item> h2</Grid>
                        <Grid item> h3</Grid>
                    </Grid>
                </Grid>
            </Container>
        </div>
    )
}

export default Home;
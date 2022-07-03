use futures::prelude::*;

use tokio::net::{UnixListener, UnixStream};

use tokio_yamux::{config::Config, session::Session};

mod proto;
use proto::{api, api_ttrpc};

use ttrpc::context::{self, Context};
use ttrpc::Client;

use std::fs;
use std::io::Result;
use std::path::Path;
use std::sync::mpsc::channel;
use std::thread;

fn main() {
    remove_if_sock_exist("/tmp/benchmark-yamux-proxy.sock").unwrap();

    let (send, recv) = channel();

    thread::spawn(move || {
        let mut rt = tokio::runtime::Runtime::new().unwrap();

        rt.block_on(async move {
            let socket = UnixStream::connect("/tmp/benchmark-yamux-server.sock")
                .await
                .unwrap();

            let mut session = Session::new_client(socket, Config::default());
            let mut control = session.control();

            tokio::spawn(async move {
                loop {
                    match session.next().await {
                        Some(Ok(_)) => (),
                        Some(Err(e)) => {
                            print!("{}", e);
                        }
                        None => {
                            print!("closed");
                        }
                    }
                }
            });

            let mut listener = UnixListener::bind("/tmp/benchmark-yamux-proxy.sock").unwrap();

            send.send("proxy listener is already, go").unwrap();
            while let Ok((inbound, _)) = listener.accept().await {
                let outbound = control.open_stream().await.unwrap();

                let (mut ri, mut wi) = inbound.into_split();
                let (mut ro, mut wo) = tokio::io::split(outbound);

                let client_to_server =
                    tokio::spawn(async move { tokio::io::copy(&mut ri, &mut wo).await.unwrap() });

                let server_to_client =
                    tokio::spawn(async move { tokio::io::copy(&mut ro, &mut wi).await.unwrap() });

                tokio::spawn(async move {
                    tokio::try_join!(client_to_server, server_to_client).unwrap();
                });
            }
        });
    });

    println!("{}", recv.recv().unwrap());

    let conn = Client::connect("unix:///tmp/benchmark-yamux-proxy.sock").unwrap();
    let cli = api_ttrpc::UnknownHubClient::new(conn);

    let req = api::ReadRequest::new();
    for _ in 1..100 {
        let now = std::time::Instant::now();

        cli.read(default_ctx(), &req).unwrap();

        let du = now.elapsed();
        if du > std::time::Duration::from_millis(10) {
            println!("{:?}", du);
        }
    }
}

pub fn remove_if_sock_exist(sock_addr: &str) -> Result<()> {
    if Path::new(sock_addr).exists() {
        fs::remove_file(&sock_addr)?;
    }
    Ok(())
}

fn default_ctx() -> Context {
    let ctx = context::with_timeout(0);
    ctx
}

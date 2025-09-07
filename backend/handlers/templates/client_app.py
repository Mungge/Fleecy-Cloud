"""flower-demo: A Flower / PyTorch app."""

import argparse
import os
import torch

import flwr as fl
from flwr.client import ClientApp, NumPyClient
from flwr.common import Context
from task import Net, get_weights, load_data, set_weights, test, train


# Define Flower Client and client_fn
class FlowerClient(NumPyClient):
    def __init__(self, net, trainloader, valloader, local_epochs):
        self.net = net
        self.trainloader = trainloader
        self.valloader = valloader
        self.local_epochs = local_epochs
        self.device = torch.device("cuda:0" if torch.cuda.is_available() else "cpu")
        self.net.to(self.device)

    def fit(self, parameters, config):
        set_weights(self.net, parameters)
        train_loss = train(
            self.net,
            self.trainloader,
            self.local_epochs,
            self.device,
        )
        return (
            get_weights(self.net),
            len(self.trainloader.dataset),
            {"train_loss": train_loss},
        )

    def evaluate(self, parameters, config):
        set_weights(self.net, parameters)
        loss, metrics = test(self.net, self.valloader, self.device)
        
        num_examples = metrics.pop("num_examples", len(self.valloader.dataset))
        return loss, num_examples, metrics


def client_fn(context: Context):
    # Load model and data
    net = Net()
    partition_id = context.node_config["partition-id"]
    num_partitions = context.node_config["num-partitions"]
    trainloader, valloader = load_data(partition_id, num_partitions)
    local_epochs = context.run_config["local-epochs"]

    # Return Client instance
    return FlowerClient(net, trainloader, valloader, local_epochs).to_client()


# Flower ClientApp (for flwr run command)
app = ClientApp(client_fn)


# 직접 실행을 위한 메인 함수
def main():
    parser = argparse.ArgumentParser(description="Flower Client")
    parser.add_argument("--server-address", default="localhost:9092",
                       help="Server address (default: localhost:9092)")
    parser.add_argument("--partition-id", type=int, default=0,
                       help="Client partition ID (default: 0)")
    parser.add_argument("--num-partitions", type=int, default=2,
                       help="Total number of partitions (default: 2)")
    parser.add_argument("--local-epochs", type=int, default=1,
                       help="Number of local epochs (default: 1)")
    
    args = parser.parse_args()
    
    print(f"=== Flower Client Configuration ===")
    print(f"Server address: {args.server_address}")
    print(f"Partition ID: {args.partition_id}")
    print(f"Total partitions: {args.num_partitions}")
    print(f"Local epochs: {args.local_epochs}")
    print(f"===================================")
    
    # Load model and data
    net = Net()
    trainloader, valloader = load_data(args.partition_id, args.num_partitions)
    
    # Create client instance
    client = FlowerClient(net, trainloader, valloader, args.local_epochs)
    
    # Start client
    print("Starting Flower client...")
    fl.client.start_client(
        server_address=args.server_address,
        client=client,
    )


if __name__ == "__main__":
    main()

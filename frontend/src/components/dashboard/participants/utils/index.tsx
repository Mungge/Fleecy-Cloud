import { Badge } from "@/components/ui/badge";

export const getStatusBadge = (status: string) => {
  const colorClass =
    status === "active"
      ? "bg-green-500"
      : status === "inactive"
      ? "bg-gray-500"
      : "bg-yellow-500";
  return <Badge className={colorClass}>{status}</Badge>;
};

export const getVMStatusBadge = (status: string) => {
  return (
    <Badge
      className={
        status === "ACTIVE"
          ? "bg-green-500"
          : status === "SHUTOFF"
          ? "bg-gray-500"
          : status === "ERROR" || status == "INACTIVE"
          ? "bg-red-500"
          : "bg-yellow-500"
      }
    >
      {status}
    </Badge>
  );
};

export const formatBytes = (bytes: number) => {
  if (bytes === 0) return "0 GB";
  const gb = bytes / 1024;
  return `${gb.toFixed(1)} GB`;
};

export const getLastAddress = (
  addresses: Record<string, Array<{ addr: string; type: string }>>
) => {
  if (!addresses || Object.keys(addresses).length === 0) return null;

  const allAddresses = Object.entries(addresses).flatMap(
    ([networkName, addressList]: [
      string,
      Array<{ addr: string; type: string }>
    ]) =>
      addressList.map((addr) => ({
        ...addr,
        networkName,
      }))
  );

  return allAddresses[allAddresses.length - 1];
};

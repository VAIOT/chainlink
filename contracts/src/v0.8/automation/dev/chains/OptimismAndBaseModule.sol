// SPDX-License-Identifier: BUSL-1.1
pragma solidity 0.8.16;

import {OVM_GasPriceOracle} from "./../../../vendor/@eth-optimism/contracts/v0.8.9/contracts/L2/predeploys/OVM_GasPriceOracle.sol";
import "../interfaces/v2_2/IChainSpecific.sol";

contract OptimismAndBaseModule is IChainSpecific {
  /// @dev OP_L1_DATA_FEE_PADDING includes 35 bytes for L1 data padding for Optimism and BASE
  bytes internal constant OP_L1_DATA_FEE_PADDING =
    hex"ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff";
  /// @dev OVM_GASPRICEORACLE_ADDR is the address of the OVM_GasPriceOracle precompile on Optimism.
  /// @dev reference: https://community.optimism.io/docs/developers/build/transaction-fees/#estimating-the-l1-data-fee
  address private constant OVM_GASPRICEORACLE_ADDR = address(0x420000000000000000000000000000000000000F);
  OVM_GasPriceOracle private constant OVM_GASPRICEORACLE = OVM_GasPriceOracle(OVM_GASPRICEORACLE_ADDR);

  function blockHash(uint256 blockNumber) external view returns (bytes32) {
    return blockhash(blockNumber);
  }

  function blockNumber() external view returns (uint256) {
    return block.number;
  }

  function getL1Fee(bytes calldata txCallData) external view returns (uint256) {
    return OVM_GASPRICEORACLE.getL1Fee(bytes.concat(txCallData, OP_L1_DATA_FEE_PADDING));
  }

  function getMaxL1Fee(uint256 dataSize) external view returns (uint256) {
    // fee is 4 per 0 byte, 16 per non-zero byte. Worst case we can have all non zero-bytes.
    // Instead of setting bytes to non-zero, we initialize 'new bytes' of length 4*dataSize to cover for zero bytes.
    bytes memory txCallData = new bytes(4 * dataSize);
    return OVM_GASPRICEORACLE.getL1Fee(bytes.concat(txCallData, OP_L1_DATA_FEE_PADDING));
  }
}

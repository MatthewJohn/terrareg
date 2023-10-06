
class ResultData:
    """Object containing search results."""

    @property
    def rows(self):
        """Return data rows."""
        return self._rows

    @property
    def count(self):
        """Return count."""
        return self._count

    @property
    def meta(self):
        """Return API meta for limit/offsets."""
        # Setup base metadata with current offset and limit
        meta_data = {
            "limit": self._limit,
            "current_offset": self._offset,
        }

        # If current offset is not 0,
        # Set previous offset as current offset minus the current limit,
        # or 0, depending on whichever is higher.
        if self._offset > 0:
            meta_data['prev_offset'] = (self._offset - self._limit) if (self._offset >= self._limit) else 0

        # If the current count of results is greater than the next offset,
        # provide the next offset in the metadata
        next_offset = (self._offset + self._limit)
        if self.count > next_offset:
            meta_data['next_offset'] = next_offset

        return meta_data

    def __init__(self, offset: int, limit: int, rows: list, count: str):
        """Store member variables"""
        self._offset = offset
        self._limit = limit
        self._rows = rows
        self._count = count
